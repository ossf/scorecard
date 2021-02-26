package e2e

import (
	"io/ioutil"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("E2E TEST:Disk Cache", func() {
	Context("E2E TEST:Disk cache is caching the request", func() {
		It("Should cache requests to the disk", func() {
			const (
				UseDiskCache  = "USE_DISK_CACHE"
				DiskCachePath = "DISK_CACHE_PATH"
			)
			Expect(len(os.Getenv(UseDiskCache))).ShouldNot(BeZero())
			Expect(len(os.Getenv(DiskCachePath))).ShouldNot(BeZero())

			// Validating if the files have been cached.
			files, err := ioutil.ReadDir(path.Join("../", os.Getenv(DiskCachePath)))

			Expect(err).Should(BeNil())
			Expect(len(files)).Should(BeNumerically(">", 1), "Number of files should be greater than 1")
		})
	})
})
