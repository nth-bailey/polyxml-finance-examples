cmake_minimum_required(VERSION 3.20)
project(fednow_pacs008_adapter_cpp CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

add_executable(fednow_pacs008_adapter main.cpp)

target_include_directories(fednow_pacs008_adapter PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/../../generated/cpp
)
