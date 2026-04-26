import re

with open('ability/handlers.go', 'r') as f:
    content = f.read()

pattern = r'insertApiLog\(([^)]+)\)'

def fix_match(match):
    args = match.group(1).split(',')
    if len(args) == 10:
        new_args = args[:9] + [' c.Query("carrier_agent_uuid")'] + args[9:]
        return 'insertApiLog(' + ', '.join(new_args) + ')'
    return match.group(0)

content = re.sub(pattern, fix_match, content)

with open('ability/handlers.go', 'w') as f:
    f.write(content)

print("Fixed insertApiLog calls")
