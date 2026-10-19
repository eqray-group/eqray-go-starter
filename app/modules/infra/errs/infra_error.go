// Package errs 定义 infra 模块的业务错误码，编码规则见 docs 8.3(2)。
package errs

import baseErrs "eqray-go-starter/app/modules/base/errs"

// infra 模块业务错误码：1-001-{子模块}-{错误码}，file 占用 002 子模块。
var (
	InfraFileUploadFailed   = baseErrs.ErrorCode{Code: 1_001_002_001, Msg: "文件上传失败"}
	InfraFileTypeNotAllowed = baseErrs.ErrorCode{Code: 1_001_002_002, Msg: "不支持的文件类型，仅支持 jpg / jpeg / png / gif / webp / bmp"}
	InfraFileTooLarge       = baseErrs.ErrorCode{Code: 1_001_002_003, Msg: "文件不能超过 %vMB"}
	InfraFileNotExists      = baseErrs.ErrorCode{Code: 1_001_002_004, Msg: "文件【%v】不存在"}

	InfraFileConfigNotExists       = baseErrs.ErrorCode{Code: 1_001_003_001, Msg: "文件配置【%v】不存在"}
	InfraFileConfigMasterForbidden = baseErrs.ErrorCode{Code: 1_001_003_002, Msg: "主配置【%v】不允许删除"}
	InfraStorageNotSupported       = baseErrs.ErrorCode{Code: 1_001_003_003, Msg: "暂不支持的存储类型【%v】"}

	InfraConfigNotExists        = baseErrs.ErrorCode{Code: 1_001_001_001, Msg: "参数【%v】不存在"}
	InfraConfigKeyExists        = baseErrs.ErrorCode{Code: 1_001_001_002, Msg: "参数键【%v】已存在"}
	InfraConfigBuiltinForbidden = baseErrs.ErrorCode{Code: 1_001_001_003, Msg: "系统内置参数【%v】不允许删除"}
)
