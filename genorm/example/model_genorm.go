// Generated with genorm. DO NOT EDIT.
package example

import (
	"database/sql"
)

const (
	selectPatientViewByIdSql = "SELECT id, doctor_id, patient_id, pulse_type_id, diagnosis, doctor_name, pulse_type_name FROM vw_patients WHERE id = $1"
	selectPatientByIdSql     = "SELECT id, doctor_id, patient_id, pulse_type_id, diagnosis FROM patients WHERE id = $1"
	insertPatientSql         = "INSERT INTO patients ( doctor_id, patient_id, pulse_type_id, diagnosis ) VALUES ( $1, $2, $3, $4 ) RETURNING id"
	updatePatientSql         = "WITH rows AS (UPDATE patients SET doctor_id = $1, patient_id = $2, pulse_type_id = $3, diagnosis = $4 WHERE id = $5 RETURNING 1) SELECT count(*) FROM rows"
)

var (
	selectPatientViewByIdStmt *sql.Stmt
	selectPatientByIdStmt     *sql.Stmt
	insertPatientStmt         *sql.Stmt
	updatePatientStmt         *sql.Stmt
)

func getPatientViewById(id int64) (PatientView, error) {
	t := PatientView{}

	err := selectPatientViewByIdStmt.QueryRow(id).Scan(
		&t.Id,
		&t.DoctorId,
		&t.PatientId,
		&t.PulseTypeId,
		&t.Diagnosis,
		&t.DoctorName,
		&t.PulseTypeName)
	if err != nil {
		return t, err
	}

	return t, nil
}

func getPatientById(id int64) (Patient, error) {
	t := Patient{}

	err := selectPatientByIdStmt.QueryRow(id).Scan(
		&t.Id,
		&t.DoctorId,
		&t.PatientId,
		&t.PulseTypeId,
		&t.Diagnosis)
	if err != nil {
		return t, err
	}

	return t, nil
}

func insertPatient(tx *sql.Tx, t *Patient) error {
	row := tx.Stmt(insertPatientStmt).QueryRow(
		t.DoctorId,
		t.PatientId,
		t.PulseTypeId,
		t.Diagnosis)
	return row.Scan(&t.Id)
}

func updatePatient(tx *sql.Tx, t *Patient) (int, error) {
	row := tx.Stmt(updatePatientStmt).QueryRow(
		t.DoctorId,
		t.PatientId,
		t.PulseTypeId,
		t.Diagnosis,
		t.Id)
	var cnt int
	if err := row.Scan(&cnt); err != nil {
		return 0, err
	}
	return cnt, nil
}
