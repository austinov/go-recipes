package example

//genorm:vw_patients:view
type PatientView struct {
	Patient       `genorm:",embed"`
	DoctorName    string `genorm:"doctor_name"`
	PulseTypeName string `genorm:"pulse_type_name"`
}

//genorm:patients
type Patient struct {
	Id          int64  `genorm:"id,pk"`
	DoctorId    int32  `genorm:"doctor_id"`
	PatientId   int32  `genorm:"patient_id"`
	PulseTypeId int32  `genorm:"pulse_type_id"`
	Diagnosis   string `genorm:"diagnosis"`
}
