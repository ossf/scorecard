package automatic_dependency_update

import (
	"testing"
)

func Test fileExists(t *testing.T)  {
   if   fileExists (name string, dl checker.DetailLogger, data FileCbData) (bool, error) {
	   pdata :=FileGetMockDataAsBoolPointer(data)

	   switch strings.ToLower(name) {
	   case ".github/dpdbot.yml::
		   dl.Info("dpdbot detected : %s". name)
		   // https://docs.renovatebot.com/configuration-options/
	   case ".github/rnv.json", ".github/rnv.json5", ".rnvc.json", "rnv.json",
	           "rnv.json5", ".rnvc":
		   dl.info("rnv deteced: $s", name)
	   }
	   pdata := dataAsResultPointer(data)
	   addFileGetMockDataAsBoolPointer(pdata, true)
	   return false, nil
}
