mod search;
mod template;

use std::{
    ffi::{CStr, CString},
    os::raw::c_char,
    panic::{catch_unwind, AssertUnwindSafe},
};

fn input(ptr: *const c_char) -> Result<String, String> {
    if ptr.is_null() {
        return Err("null input".into());
    }
    // The C caller owns this NUL-terminated UTF-8 buffer for the duration of the call.
    unsafe { CStr::from_ptr(ptr) }
        .to_str()
        .map(str::to_owned)
        .map_err(|e| e.to_string())
}

fn output(value: Result<String, String>) -> *mut c_char {
    let value = value.unwrap_or_else(|e| serde_json::json!({"error": e}).to_string());
    CString::new(value)
        .map(CString::into_raw)
        .unwrap_or(std::ptr::null_mut())
}

fn guard(call: impl FnOnce() -> Result<String, String>) -> *mut c_char {
    output(catch_unwind(AssertUnwindSafe(call)).unwrap_or_else(|_| Err("engine panic".into())))
}

#[no_mangle]
pub extern "C" fn snipbox_search(
    commands_json: *const c_char,
    query: *const c_char,
) -> *mut c_char {
    guard(|| search::search(&input(commands_json)?, &input(query)?))
}

#[no_mangle]
pub extern "C" fn snipbox_parse_template(command_json: *const c_char) -> *mut c_char {
    guard(|| template::parse(&input(command_json)?))
}

#[no_mangle]
pub extern "C" fn snipbox_render_template(
    command: *const c_char,
    values_json: *const c_char,
) -> *mut c_char {
    guard(|| template::render(&input(command)?, &input(values_json)?))
}

#[no_mangle]
pub extern "C" fn snipbox_free_string(value: *mut c_char) {
    if !value.is_null() {
        unsafe {
            drop(CString::from_raw(value));
        }
    }
}
