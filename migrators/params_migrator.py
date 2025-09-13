import json
import sys
from pathlib import Path

def transform_param(param):
    """Transform a single param object to the new format."""
    if not all(k in param for k in ("key", "type", "value")):
        return param  # leave untouched if missing expected fields
    
    return {
        "type": "EQUALS",
        "left": {
            "type": param["type"],  # BODY / QUERY / PATH
            "id": param["key"]
        },
        "right": {
            "type": "STRING",
            "value": param["value"]
        }
    }

def transform_structure(obj):
    """Recursively traverse JSON structure and transform 'params' lists."""
    if isinstance(obj, dict):
        new_obj = {}
        for k, v in obj.items():
            if k == "params" and isinstance(v, list):
                new_obj[k] = [transform_param(p) for p in v]
            else:
                new_obj[k] = transform_structure(v)
        return new_obj
    elif isinstance(obj, list):
        return [transform_structure(item) for item in obj]
    else:
        return obj

def main(input_file, output_file):
    # Read input JSON
    with open(input_file, "r", encoding="utf-8") as f:
        data = json.load(f)

    # Transform
    transformed = transform_structure(data)

    # Write output JSON (pretty-printed)
    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(transformed, f, indent=2, ensure_ascii=False)

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print(f"Usage: python {Path(__file__).name} input.json output.json")
        sys.exit(1)
    
    main(sys.argv[1], sys.argv[2])
