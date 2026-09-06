import unittest,tempfile,pathlib,io,json,tarfile,types,os
from unittest.mock import patch
source=(pathlib.Path(__file__).resolve().parents[1]/'backup/export-router-production.py').read_text()
class ExportTests(unittest.TestCase):
 def export(self,mode):
  with tempfile.TemporaryDirectory() as d:
   code=source.replace("pathlib.Path('/home/nio/backups/fx-production')",'pathlib.Path('+repr(d)+')')
   calls=[];output=io.BytesIO()
   def run(args,**kw):
    calls.append(args)
    out=kw.get('stdout')
    if hasattr(out,'write'):out.write(b'fixture data')
    return types.SimpleNamespace(returncode=0)
   meta=[{'Name':'/fx-production-'+n,'Image':'sha256:test'} for n in ['router','studio','postgres','redis']]
   with patch.dict(os.environ,{'SSH_ORIGINAL_COMMAND':mode}),patch('subprocess.run',side_effect=run),patch('subprocess.check_output',return_value=json.dumps(meta).encode()),patch('sys.stdout',types.SimpleNamespace(buffer=output)):
    namespace={}
    exec(compile(code,'export-test','exec'),namespace)
    namespace['lock'].close()
   output.seek(0)
   with tarfile.open(fileobj=output) as t:
    names=t.getnames();manifest=json.load(t.extractfile('manifest.json'))
   return calls,names,manifest
 def test_studio_only_does_not_dump_router(self):
  calls,names,manifest=self.export('backup-studio-v1')
  dumps=[a[a.index('-d')+1] for a in calls if 'pg_dump' in a]
  self.assertEqual(dumps,['nanafox_studio_prod']);self.assertNotIn('sub2api.dump',names)
  self.assertEqual(manifest['databases'],['nanafox_studio_prod'])
  self.assertEqual(manifest['source'],'studio-production')
  archive=next(a for a in calls if 'tar' in a)
  self.assertIn('recovery/studio.production.env',archive)
  self.assertNotIn('recovery/router.preview.json',archive)
 def test_existing_combined_mode_preserved(self):
  calls,names,manifest=self.export('backup-v1')
  self.assertIn('sub2api.dump',names);self.assertIn('nanafox_studio_prod.dump',names)
  self.assertEqual(manifest['databases'],['sub2api','nanafox_studio_prod'])
 def test_other_ssh_commands_rejected_before_subprocess(self):
  with patch.dict(os.environ,{'SSH_ORIGINAL_COMMAND':'shell'}),patch('subprocess.run') as run:
   with self.assertRaises(SystemExit):exec(compile(source,'export-test','exec'),{})
   run.assert_not_called()
if __name__=='__main__':unittest.main()
