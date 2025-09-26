import os
import sys

# Tambahkan folder msys64 runtime ke PATH Python
dll_path = r"C:\msys64\ucrt64\bin"
os.add_dll_directory(dll_path)

print("🔍 DLL path added:", dll_path)

try:
    import capytaine.green_functions.libs.Delhommeau_float64
    print("✅ Delhommeau_float64 loaded successfully")
except Exception as e:
    print("❌ Failed to load Delhommeau_float64:", e)

try:
    import capytaine.green_functions.libs.Delhommeau_float32
    print("✅ Delhommeau_float32 loaded successfully")
except Exception as e:
    print("❌ Failed to load Delhommeau_float32:", e)
