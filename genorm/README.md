# genorm 

genorm is a simple utility to generate of dao code by model structs.

It was inspired by awesome [reform](https://github.com/go-reform/reform) and uses some pieces of the code.
Genorm id much simpler than reform and it's made for specific task in another project.
Genorm generates code for PostgreSQL and supported embed structs.

To use:

1. Build genorm:
```
  $ go build
```

2. Define a model – `struct` representing a table or view. For example, in file `model.go`:

```go
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
```

3. Run:

```
  $ genorm [flags] [source directories] 
```

Flags:
  - dst-path - destination path to store files
  - dst-pack - destination package name

For example, after run `./genorm ./example` result of code generation for `./example/model.go` is presented in `./example/model_genorm.go`.
