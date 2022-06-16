package internal_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/paketo-buildpacks/cpython/cmd/env/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/v2/matchers"
)

func testRun(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("when $PYTHONPYCACHEPREFIX is not set", func() {
		it("sets it to $HOME/.pycache", func() {
			buffer := bytes.NewBuffer(nil)
			err := internal.Run([]string{"HOME=/some-home"}, buffer)
			Expect(err).NotTo(HaveOccurred())

			Expect(buffer.String()).To(MatchTOML(`
				PYTHONPYCACHEPREFIX = "/some-home/.pycache"
			`))
		})
	})

	context("when $PYTHONPYCACHEPREFIX is already set", func() {
		it("preserves the existing value", func() {
			buffer := bytes.NewBuffer(nil)
			err := internal.Run([]string{
				"HOME=/some-home",
				"PYTHONPYCACHEPREFIX=/other-home/.pycache",
			}, buffer)
			Expect(err).NotTo(HaveOccurred())

			Expect(buffer.String()).To(BeEmpty())
		})
	})

	context("failure cases", func() {
		context("when the output cannot be written to", func() {
			var file *os.File

			it.Before(func() {
				file, err := os.CreateTemp("", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(file.Close()).To(Succeed())
				Expect(os.Remove(file.Name())).To(Succeed())
			})

			it("returns an error", func() {
				err := internal.Run([]string{"HOME=/some-home"}, file)
				Expect(err).To(MatchError("invalid argument"))
			})
		})
	})
}
