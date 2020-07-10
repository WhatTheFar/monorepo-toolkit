package presenter

import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/whatthefar/monorepo-toolkit/pkg/interactor"
)

func TestBuildProjectsPresenter(t *testing.T) {
	Convey("Given a buildProjectsPresenter", t, func() {
		var buf bytes.Buffer
		p := &buildProjectsPresenter{writer: &buf}

		Convey("When calls BuildSkippedFor", func() {
			infos := []*interactor.BuildInfo{
				{ProjectName: "app1", BuildID: "123"},
				{ProjectName: "app2", BuildID: "456"},
			}
			p.WaitingFor(infos)
			got := buf.String()

			Convey("It should print a correct string", func() {
				want := `Waiting for build app1(123) app2(456)...
`
				So(got, ShouldEqual, want)
			})
		})
	})
}
