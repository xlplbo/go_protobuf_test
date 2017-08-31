:: 一键编译所有proto文件
@echo off
@echo start compile proto file ...
.\protoc\bin\protoc.exe --plugin=protoc-gen-go=.\protoc\bin\protoc-gen-go.exe --go_out . *.proto
if %errorlevel% == 0 ( 
    @echo compile success! 
) else ( 
    @echo compile failed! 
)