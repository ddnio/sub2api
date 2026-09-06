# Deployed program reference, 2026-09-07. Paths retain actual runtime identifiers.
import os,pathlib,subprocess,tempfile,hashlib,json,tarfile,fcntl,datetime,shlex,argparse
parser=argparse.ArgumentParser();parser.add_argument('--studio',action='store_true');studio_only=parser.parse_args().studio
os.umask(0o077)
identity_root=pathlib.Path('/home/Nio/.local/share/nanafox-fx-production-backup')
root=identity_root/'studio' if studio_only else identity_root
root.mkdir(mode=0o700,parents=True,exist_ok=True)
lock=(root/'run.lock').open('w')
try:fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
except BlockingIOError:
 print(json.dumps({'stage':'already_running'}),flush=True);raise SystemExit(0)
ssh=['ssh','-i',str(identity_root/'id_ed25519'),'-o','IdentitiesOnly=yes','-o','BatchMode=yes','-o','StrictHostKeyChecking=yes','-o','UserKnownHostsFile='+str(identity_root/'known_hosts'),'-o','ConnectTimeout=15','-o','ServerAliveInterval=30','-o','ServerAliveCountMax=6','nio@43.106.8.109','backup-studio-v1' if studio_only else 'backup-v1']
def event(**data):print(json.dumps(data),flush=True)
def mc(command,**kw):
 auth='export MC_HOST_backup="http://$MINIO_ROOT_USER:$MINIO_ROOT_PASSWORD@127.0.0.1:9000"; '
 return subprocess.Popen(['docker','exec','-i','nanafox-backup-minio','sh','-ec',auth+'mc '+command],**kw)
try:
 with tempfile.TemporaryDirectory(prefix='run-',dir=root) as temp:
  p=pathlib.Path(temp);bundle=p/'bundle.tar';event(stage='export_started',time_utc=datetime.datetime.now(datetime.timezone.utc).isoformat())
  with bundle.open('wb') as out:
   proc=subprocess.run(ssh,stdout=out,stderr=subprocess.PIPE)
  if proc.returncode:raise RuntimeError('Router production backup export failed; ssh exit '+str(proc.returncode))
  with tarfile.open(bundle) as archive:
   members=archive.getmembers()
   allowed={'sub2api.dump','nanafox_studio_prod.dump','globals.private.sql','containers.private.json','production-configs.private.tar.gz','manifest.json'}
   if studio_only:allowed.remove('sub2api.dump')
   assert len(members)==len(allowed) and {m.name for m in members}==allowed
   for m in members:
    assert m.isfile() and '/' not in m.name
    with archive.extractfile(m) as inp,(p/m.name).open('wb') as out:
     for b in iter(lambda:inp.read(4*1024*1024),b''):out.write(b)
  bundle.unlink();manifest=json.loads((p/'manifest.json').read_text());stamp=manifest['timestamp_utc']
  assert len(stamp)==16 and stamp.endswith('Z')
  assert {e['name'] for e in manifest['files']}==allowed-{'manifest.json'}
  assert len(manifest['files'])==len(allowed)-1
  assert manifest['databases']==(['nanafox_studio_prod'] if studio_only else ['sub2api','nanafox_studio_prod'])
  prefix='backup/nanafox-postgres-backups/'+('studio-production/' if studio_only else 'fx-production/')+stamp+'/'
  results=[]
  for entry in manifest['files']+[{'name':'manifest.json'}]:
   file=p/entry['name'];h=hashlib.sha256()
   with file.open('rb') as inp:
    for b in iter(lambda:inp.read(4*1024*1024),b''):h.update(b)
   digest=h.hexdigest();size=file.stat().st_size
   assert 'sha256' not in entry or entry['sha256']==digest
   assert 'bytes' not in entry or entry['bytes']==size
   with file.open('rb') as inp:
    upload=mc('pipe '+shlex.quote(prefix+file.name),stdin=inp,stdout=subprocess.DEVNULL,stderr=subprocess.PIPE);_,err=upload.communicate()
   if upload.returncode:raise RuntimeError('MinIO upload failed: '+file.name)
   read=mc('cat '+shlex.quote(prefix+file.name),stdin=subprocess.DEVNULL,stdout=subprocess.PIPE,stderr=subprocess.PIPE);h=hashlib.sha256();n=0
   for b in iter(lambda:read.stdout.read(4*1024*1024),b''):h.update(b);n+=len(b)
   _,err=read.communicate()
   assert read.returncode==0 and n==size and h.hexdigest()==digest,'MinIO readback verification failed'
   results.append({'name':file.name,'bytes':size,'sha256':digest,'verified':True});event(stage='file_verified',file=file.name,bytes=size)
  result={'stage':'complete','timestamp_utc':stamp,'minio_prefix':prefix.removeprefix('backup/'),'databases':manifest['databases'],'files':results,'completed_utc':datetime.datetime.now(datetime.timezone.utc).isoformat()}
  (root/'last-success.json').write_text(json.dumps(result,indent=2));event(**result)
except Exception as ex:
 result={'stage':'failed','error_type':type(ex).__name__,'time_utc':datetime.datetime.now(datetime.timezone.utc).isoformat()}
 (root/'last-failure.json').write_text(json.dumps(result));event(**result);raise
