// TODO: move to shared library
export interface RequiredInputDTO {
  id: string
  coursePhaseTypeID: string
  dtoName: string
  specification: { [key: string]: string }
  optional: boolean
}
