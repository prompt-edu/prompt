package developerProfileDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
)

type DeviceWithParticipationID struct {
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Devices               []string  `json:"devices"`
}

func getDeviceDTOFromDBModel(deviceWithParticipationID db.GetDevicesForCoursePhaseRow) DeviceWithParticipationID {
	return DeviceWithParticipationID{
		CourseParticipationID: deviceWithParticipationID.CourseParticipationID,
		Devices:               deviceWithParticipationID.Devices,
	}
}

func GetDeviceWithParticipationIDFromDBModel(devicesWithParticipationID []db.GetDevicesForCoursePhaseRow) []DeviceWithParticipationID {
	deviceDTOs := make([]DeviceWithParticipationID, 0, len(devicesWithParticipationID))
	for _, deviceWithParticipationID := range devicesWithParticipationID {
		deviceDTO := getDeviceDTOFromDBModel(deviceWithParticipationID)
		deviceDTOs = append(deviceDTOs, deviceDTO)
	}
	return deviceDTOs
}
