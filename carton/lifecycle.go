/*
** copyright [2013-2016] [Megam Systems]
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
** http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
 */

package carton

import (
	"fmt"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
	constants "github.com/virtengine/libgo/utils"
	lw "github.com/virtengine/libgo/writer"
	"github.com/virtengine/vertice/provision"
)

type LifecycleOpts struct {
	B         *provision.Box
	start     time.Time
	Hard      bool
	logWriter lw.LogWriter
	writer    io.Writer
}

func (cy *LifecycleOpts) setLogger() {
	cy.start = time.Now()
	cy.logWriter = lw.NewLogWriter(cy.B)
	cy.writer = io.MultiWriter(&cy.logWriter)
}

//if the state is in running, started, stopped, restarted then allow it to be lcycled.
// to-do: allow states that ends with "*ing or *ed" that should fix this generically.
func (cy *LifecycleOpts) canCycleStart() bool {
	return cy.B.State == constants.StateStopped
}

func (cy *LifecycleOpts) canCycleStop() bool {
	return cy.B.State == constants.StateRunning ||
		cy.B.State == constants.StatePostError
}

func (cy *LifecycleOpts) process(process string) string {
	if cy.Hard {
		return "hard-" + process
	}
	return process
}

// Starts  the box.
func Start(cy *LifecycleOpts) error {
	log.Debugf("  start cycle for box (%s, %s)", cy.B.Id, cy.B.GetFullName())
	cy.setLogger()
	defer cy.logWriter.Close()
	if cy.canCycleStart() {
		if err := ProvisionerMap[cy.B.Provider].Start(cy.B, cy.process(constants.START), cy.writer); err != nil {
			return err
		}
	} else {
		fmt.Printf("start (%s, %s, %s) Unsuccessfull because of lifecycle not allowed\n", cy.B.GetFullName(), cy.B.Status.String(), time.Since(cy.start))
	}
	fmt.Fprintf(cy.writer, "    start (%s, %s, %s) OK\n", cy.B.GetFullName(), cy.B.Status.String(), time.Since(cy.start))
	return nil
}

// Stops the box
func Stop(cy *LifecycleOpts) error {
	log.Debugf("  stop cycle for box (%s, %s)", cy.B.Id, cy.B.GetFullName())
	cy.setLogger()
	defer cy.logWriter.Close()
	if cy.canCycleStop() {

		if err := ProvisionerMap[cy.B.Provider].Stop(cy.B, cy.process(constants.STOP), cy.writer); err != nil {
			return err
		}
	} else {
		fmt.Printf("start (%s, %s, %s) Unsuccessfull because of lifecycle not allowed\n", cy.B.GetFullName(), cy.B.Status.String(), time.Since(cy.start))
	}
	fmt.Fprintf(cy.writer, "    stop (%s, %s, %s) OK\n", cy.B.GetFullName(), cy.B.Status.String(), time.Since(cy.start))
	return nil
}

// Restart the box.
func Restart(cy *LifecycleOpts) error {
	log.Debugf("  restart cycle for box (%s, %s)", cy.B.Id, cy.B.GetFullName())
	cy.setLogger()
	defer cy.logWriter.Close()
	if cy.canCycleStop() {
		if err := ProvisionerMap[cy.B.Provider].Restart(cy.B, cy.process(constants.RESTART), cy.writer); err != nil {
			return err
		}
	} else {
		fmt.Printf("start (%s, %s, %s) Unsuccessfull because of lifecycle not allowed\n", cy.B.GetFullName(), cy.B.Status.String(), time.Since(cy.start))
	}
	fmt.Fprintf(cy.writer, "    restart (%s, %s, %s) OK\n", cy.B.GetFullName(), cy.B.Status.String(), time.Since(cy.start))
	return nil
}

// Stops the box
func SuspendBox(cy *LifecycleOpts) error {
	log.Debugf("  suspend cycle for box (%s, %s)", cy.B.Id, cy.B.GetFullName())
	cy.setLogger()
	defer cy.logWriter.Close()
	if cy.canCycleStop() {

		if err := ProvisionerMap[cy.B.Provider].Suspend(cy.B, cy.process(constants.SUSPEND), cy.writer); err != nil {
			return err
		}
	} else {
		fmt.Printf("start (%s, %s, %s) Unsuccessfull because of lifecycle not allowed\n", cy.B.GetFullName(), cy.B.Status.String(), time.Since(cy.start))
	}
	fmt.Fprintf(cy.writer, "    suspend (%s, %s, %s) OK\n", cy.B.GetFullName(), cy.B.Status.String(), time.Since(cy.start))
	return nil
}
