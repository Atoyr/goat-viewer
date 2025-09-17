// Learn more about Tauri commands at https://tauri.app/develop/calling-rust/
use base64::Engine; // for base64 encode()
#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

#[tauri::command]
fn list_images_in_dir(dir: &str) -> Result<Vec<String>, String> {
    use std::path::Path;
    use walkdir::WalkDir;

    let exts = [
        "jpg", "jpeg", "png", "gif", "webp", "avif", "bmp",
    ];
    let root = Path::new(dir);
    if !root.exists() {
        return Err("Directory not found".into());
    }

    let mut files: Vec<(String, String)> = Vec::new();
    for entry in WalkDir::new(root).min_depth(1).max_depth(3) {
        let entry = entry.map_err(|e| e.to_string())?;
        let path = entry.path();
        if !path.is_file() { continue; }
        if let Some(ext) = path.extension().and_then(|s| s.to_str()).map(|s| s.to_lowercase()) {
            if !exts.contains(&ext.as_str()) { continue; }
        } else { continue; }

        let abs = path.to_string_lossy().into_owned();
        // Keep a stable, human-ish order: by file name (case-insensitive)
        let name = path.file_name().and_then(|s| s.to_str()).unwrap_or("").to_lowercase();
        files.push((name, abs));
    }

    files.sort_by(|a, b| a.0.cmp(&b.0));
    Ok(files.into_iter().map(|(_, p)| p).collect())
}

#[tauri::command]
fn list_images_in_zip(path: &str) -> Result<Vec<String>, String> {
    use std::fs::File;
    use std::io::Read;

    let mut file = File::open(path).map_err(|e| e.to_string())?;
    let mut data = Vec::new();
    file.read_to_end(&mut data).map_err(|e| e.to_string())?;

    let reader = std::io::Cursor::new(data);
    let mut zip = zip::ZipArchive::new(reader).map_err(|e| e.to_string())?;

    let exts = ["jpg", "jpeg", "png", "gif", "webp", "avif", "bmp"];
    let mut names: Vec<String> = Vec::new();
    for i in 0..zip.len() {
        let file = zip.by_index(i).map_err(|e| e.to_string())?;
        if file.is_dir() { continue; }
        let name = file.name().to_string();
        let has_ok_ext = name.rsplit('.').next()
            .map(|s| s.to_ascii_lowercase())
            .map(|s| exts.contains(&s.as_str()))
            .unwrap_or(false);
        if has_ok_ext { names.push(name); }
    }
    names.sort();
    Ok(names)
}

#[tauri::command]
fn read_zip_image(path: &str, entry: &str) -> Result<(String, String), String> {
    use std::fs::File;
    use std::io::Read;

    let mut file = File::open(path).map_err(|e| e.to_string())?;
    let mut data = Vec::new();
    file.read_to_end(&mut data).map_err(|e| e.to_string())?;

    let reader = std::io::Cursor::new(data);
    let mut zip = zip::ZipArchive::new(reader).map_err(|e| e.to_string())?;
    let mut f = zip.by_name(entry).map_err(|e| e.to_string())?;
    let mut buf = Vec::with_capacity(f.size() as usize);
    f.read_to_end(&mut buf).map_err(|e| e.to_string())?;

    let mime = match entry.rsplit('.').next().unwrap_or("").to_ascii_lowercase().as_str() {
        "jpg" | "jpeg" => "image/jpeg",
        "png" => "image/png",
        "gif" => "image/gif",
        "webp" => "image/webp",
        "avif" => "image/avif",
        "bmp" => "image/bmp",
        _ => "application/octet-stream",
    }.to_string();

    let b64 = base64::engine::general_purpose::STANDARD.encode(buf);
    Ok((mime, b64))
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_dialog::init())
        .invoke_handler(tauri::generate_handler![
            greet,
            list_images_in_dir,
            list_images_in_zip,
            read_zip_image
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
