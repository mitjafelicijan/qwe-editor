-----------------------------------------------------------------------
-- LEARN LUA BY EXAMPLE
-- Read from top to bottom. Run the file and modify values as you go.
-----------------------------------------------------------------------

print("Hello, Lua!")  -- Printing to the console

-----------------------------------------------------------------------
-- 1. VARIABLES & TYPES
-----------------------------------------------------------------------

local name = "Lua"        -- string
local version = 5.1       -- number (Lua only has one numeric type)
local isFun = true        -- boolean
local nothing = nil       -- nil means "no value"

print(name, version, isFun, nothing)

-----------------------------------------------------------------------
-- 2. BASIC OPERATIONS
-----------------------------------------------------------------------

local a = 10
local b = 3

print("Addition:", a + b)
print("Subtraction:", a - b)
print("Multiplication:", a * b)
print("Division:", a / b)
print("Power:", a ^ b)
print("Modulo:", a % b)

-----------------------------------------------------------------------
-- 3. STRINGS
-----------------------------------------------------------------------

local first = "Hello"
local second = "World"

-- Concatenation uses ..
local message = first .. " " .. second .. "!"
print(message)

print("Length of message:", #message)

-----------------------------------------------------------------------
-- 4. CONDITIONS (if / elseif / else)
-----------------------------------------------------------------------

local health = 25

if health > 50 then
    print("You are healthy")
elseif health > 0 then
    print("You are injured")
else
    print("You are dead")
end

-----------------------------------------------------------------------
-- 5. LOOPS
-----------------------------------------------------------------------

-- for loop
for i = 1, 5 do
    print("For loop i =", i)
end

-- while loop
local counter = 1
while counter <= 3 do
    print("While loop counter =", counter)
    counter = counter + 1
end

-----------------------------------------------------------------------
-- 6. TABLES (VERY IMPORTANT IN LUA)
-- Tables are arrays, dictionaries, objects, and structs all in one
-----------------------------------------------------------------------

-- Array-like table
local fruits = { "apple", "banana", "cherry" }

print("First fruit:", fruits[1]) -- Lua arrays start at 1

-- Loop over array
for i, fruit in ipairs(fruits) do
    print(i, fruit)
end

-- Dictionary-like table
local player = {
    name = "Hunter",
    level = 60,
    alive = true
}

print(player.name, player.level)

-----------------------------------------------------------------------
-- 7. FUNCTIONS
-----------------------------------------------------------------------

local function add(x, y)
    return x + y
end

print("Function result:", add(4, 6))

-----------------------------------------------------------------------
-- 8. FUNCTIONS AS VALUES
-----------------------------------------------------------------------

local function apply(a, b, fn)
    return fn(a, b)
end

print("Apply add:", apply(2, 3, add))

-----------------------------------------------------------------------
-- 9. SIMPLE OBJECT-LIKE TABLE
-----------------------------------------------------------------------

local Enemy = {}
Enemy.__index = Enemy

function Enemy:new(name, hp)
    local obj = {
        name = name,
        hp = hp
    }
    setmetatable(obj, self)
    return obj
end

function Enemy:takeDamage(amount)
    self.hp = self.hp - amount
    print(self.name .. " takes " .. amount .. " damage. HP =", self.hp)
end

local wolf = Enemy:new("Wolf", 40)
wolf:takeDamage(15)

-----------------------------------------------------------------------
-- 10. NIL AND CHECKING VALUES
-----------------------------------------------------------------------

local value = nil

if value == nil then
    print("Value is nil")
end

-----------------------------------------------------------------------
-- 11. GLOBAL VS LOCAL
-----------------------------------------------------------------------

globalVar = "I am global" -- avoid globals when possible!
local localVar = "I am local"

print(globalVar)
print(localVar)

-----------------------------------------------------------------------
-- 12. ERROR HANDLING
-----------------------------------------------------------------------

local function risky()
    error("Something went wrong!")
end

local success, err = pcall(risky)
print("Success:", success)
print("Error:", err)

-----------------------------------------------------------------------
-- END
-----------------------------------------------------------------------

print("You just walked through the basics of Lua 🎉")
