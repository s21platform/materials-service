# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [api/materials.proto](#api_materials-proto)
    - [ArchivedMaterialIn](#-ArchivedMaterialIn)
    - [CreateMaterialIn](#-CreateMaterialIn)
    - [CreateMaterialOut](#-CreateMaterialOut)
    - [DeleteMaterialIn](#-DeleteMaterialIn)
    - [EditMaterialIn](#-EditMaterialIn)
    - [EditMaterialOut](#-EditMaterialOut)
    - [GetAllMaterialsOut](#-GetAllMaterialsOut)
    - [GetMaterialIn](#-GetMaterialIn)
    - [GetMaterialOut](#-GetMaterialOut)
    - [Material](#-Material)
  
    - [MaterialsService](#-MaterialsService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="api_materials-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api/materials.proto



<a name="-ArchivedMaterialIn"></a>

### ArchivedMaterialIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uuid | [string](#string) |  |  |






<a name="-CreateMaterialIn"></a>

### CreateMaterialIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| title | [string](#string) |  | Заголовок материала |
| cover_image_url | [string](#string) |  | URL обложки материала |
| description | [string](#string) |  | Описание материала |
| content | [string](#string) |  | Содержимое материала |
| read_time_minutes | [int32](#int32) |  | Время чтения в минутах |






<a name="-CreateMaterialOut"></a>

### CreateMaterialOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uuid | [string](#string) |  | UUID созданного материала |






<a name="-DeleteMaterialIn"></a>

### DeleteMaterialIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uuid | [string](#string) |  |  |






<a name="-EditMaterialIn"></a>

### EditMaterialIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uuid | [string](#string) |  | UUID материала |
| title | [string](#string) |  | Заголовок материала |
| cover_image_url | [string](#string) |  | URL обложки материала |
| description | [string](#string) |  | Описание материала |
| content | [string](#string) |  | Содержание материала |
| read_time_minutes | [int32](#int32) |  | Время чтения в минутах |






<a name="-EditMaterialOut"></a>

### EditMaterialOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| material | [Material](#Material) |  | Весь материал |






<a name="-GetAllMaterialsOut"></a>

### GetAllMaterialsOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| material_list | [Material](#Material) | repeated |  |






<a name="-GetMaterialIn"></a>

### GetMaterialIn



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uuid | [string](#string) |  | UUID материала |






<a name="-GetMaterialOut"></a>

### GetMaterialOut



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| material | [Material](#Material) |  | Весь материал |






<a name="-Material"></a>

### Material



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| uuid | [string](#string) |  |  |
| owner_uuid | [string](#string) |  | UUID владельца материала |
| title | [string](#string) |  | Заголовок материала |
| cover_image_url | [string](#string) |  | URL обложки материала |
| description | [string](#string) |  | Описание материала |
| content | [string](#string) |  | Содержимое материала |
| read_time_minutes | [int32](#int32) |  | Время чтения в минутах |
| status | [string](#string) |  | Статус материала |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Время создания |
| edited_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Время последнего редактирования |
| published_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Время публикации |
| archived_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Время архивации |
| deleted_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Время удаления |
| likes_count | [int32](#int32) |  | Количество лайков |





 

 

 


<a name="-MaterialsService"></a>

### MaterialsService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateMaterial | [.CreateMaterialIn](#CreateMaterialIn) | [.CreateMaterialOut](#CreateMaterialOut) |  |
| GetMaterial | [.GetMaterialIn](#GetMaterialIn) | [.GetMaterialOut](#GetMaterialOut) |  |
| GetAllMaterials | [.google.protobuf.Empty](#google-protobuf-Empty) | [.GetAllMaterialsOut](#GetAllMaterialsOut) |  |
| EditMaterial | [.EditMaterialIn](#EditMaterialIn) | [.EditMaterialOut](#EditMaterialOut) |  |
| DeleteMaterial | [.DeleteMaterialIn](#DeleteMaterialIn) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| ArchivedMaterial | [.ArchivedMaterialIn](#ArchivedMaterialIn) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

