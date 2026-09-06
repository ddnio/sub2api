# Deployed program reference, 2026-09-07. Paths retain actual runtime identifiers.
import os,pathlib,subprocess,tempfile,hashlib,json,datetime,tarfile,sys,fcntl
os.umask(0o077)
if os.environ.get('SSH_ORIGINAL_COMMAND')!='backup-v1':
 sys.exit('Only backup-v1 is permitted')
root=pathlib.Path('/home/nio/backups/fx-production');r=pathlib.Path('/home/nio/recovery/live-production-20260906')
lock=(root/'export.lock').open('w');fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
def run(a,**kw):return subprocess.run(a,check=True,stderr=subprocess.PIPE,**kw)
with tempfile.TemporaryDirectory(prefix='run-',dir=root) as temp:
 p=pathlib.Path(temp);stamp=datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%SZ')
 for db in ['sub2api','nanafox_studio_prod']:
  with (p/(db+'.dump')).open('wb') as out:
   run(['docker','exec','fx-production-postgres','pg_dump','-U','migration_admin','-d',db,'--format=custom','--compress=6'],stdout=out)
  with (p/(db+'.dump')).open('rb') as inp:
   run(['docker','exec','-i','fx-production-postgres','pg_restore','--list'],stdin=inp,stdout=subprocess.DEVNULL)
 with (p/'globals.private.sql').open('wb') as out:run(['docker','exec','fx-production-postgres','pg_dumpall','-U','migration_admin','--globals-only'],stdout=out)
 containers=['fx-production-router','fx-production-studio','fx-production-postgres','fx-production-redis']
 meta=json.loads(subprocess.check_output(['docker','inspect']+containers));(p/'containers.private.json').write_text(json.dumps(meta))
 image=next(x['Image'] for x in meta if x['Name']=='/fx-production-postgres')
 files=['router.preview.json','router.production.json','studio.production.env','postgres.env','redis.conf','destination-secrets.private.json','site-logo.before-static.private.json']
 args=['docker','run','--rm','--network','none','--mount','type=bind,src='+str(r)+',dst=/recovery,readonly','--mount','type=bind,src=/etc/caddy,dst=/caddy,readonly','--mount','type=bind,src=/srv/nanafox/image-playground,dst=/image-playground,readonly','--mount','type=bind,src=/srv/nanafox/fx-web-assets,dst=/web-assets,readonly','--mount','type=bind,src=/srv/nanafox-landing,dst=/landing,readonly','--mount','type=bind,src=/var/lib/caddy/.local/share/caddy/certificates,dst=/certificates,readonly',image,'tar','-czf','-']
 args+=['recovery/'+f for f in files]+['caddy','image-playground','web-assets','landing','certificates/acme-v02.api.letsencrypt.org-directory/router.nanafox.com','certificates/acme-v02.api.letsencrypt.org-directory/studio.nanafox.com','certificates/acme-v02.api.letsencrypt.org-directory/nanafox.com','certificates/acme-v02.api.letsencrypt.org-directory/www.nanafox.com']
 with (p/'production-configs.private.tar.gz').open('wb') as out:run(args,stdout=out)
 records=[]
 for file in sorted(p.iterdir()):
  h=hashlib.sha256()
  with file.open('rb') as inp:
   for b in iter(lambda:inp.read(4*1024*1024),b''):h.update(b)
  records.append({'name':file.name,'bytes':file.stat().st_size,'sha256':h.hexdigest()})
 manifest={'timestamp_utc':stamp,'source':'fx-production','databases':['sub2api','nanafox_studio_prod'],'files':records,'images':{x['Name'].lstrip('/'):x['Image'] for x in meta}}
 (p/'manifest.json').write_text(json.dumps(manifest,indent=2))
 with tarfile.open(fileobj=sys.stdout.buffer,mode='w|') as archive:
  for file in sorted(p.iterdir()):archive.add(file,arcname=file.name,recursive=False)
