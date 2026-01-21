// sample.ts
// A practical TypeScript tour in one file 🚀

// --------------------
// Basic Types
// --------------------
let username: string = "Alice";
let age: number = 30;
let isAdmin: boolean = false;

let hobbies: string[] = ["gaming", "coding"];
let scores: Array<number> = [10, 20, 30];

// Tuple
let userTuple: [string, number] = ["Bob", 42];

// --------------------
// Enums
// --------------------
enum Role {
    User = "USER",
    Admin = "ADMIN",
    Moderator = "MODERATOR",
}

// --------------------
// Interfaces & Types
// --------------------
interface User {
    id: number;
    name: string;
    role: Role;
    email?: string; // optional
}

type ApiResponse<T> = {
    success: boolean;
    data: T;
    error?: string;
};

// --------------------
// Functions
// --------------------
function greet(user: User): string {
    return `Hello, ${user.name}! Your role is ${user.role}.`;
}

const add = (a: number, b: number): number => a + b;

// Function with default value
function log(message: string, level: "info" | "warn" | "error" = "info") {
    console.log(`[${level.toUpperCase()}] ${message}`);
}

// --------------------
// Classes
// --------------------
class UserService {
    private users: User[] = [];

    constructor(initialUsers: User[] = []) {
        this.users = initialUsers;
    }

    addUser(user: User): void {
        this.users.push(user);
    }

    findByRole(role: Role): User[] {
        return this.users.filter(u => u.role === role);
    }
}

// --------------------
// Generics
// --------------------
function wrapResponse<T>(data: T): ApiResponse<T> {
    return {
        success: true,
        data,
    };
}

// --------------------
// Async / Await
// --------------------
async function fetchUser(id: number): Promise<User> {
    // Fake async call
    return new Promise(resolve => {
        setTimeout(() => {
            resolve({
                id,
                name: "Charlie",
                role: Role.User,
            });
        }, 500);
    });
}

// --------------------
// Type Narrowing
// --------------------
function printValue(value: string | number) {
    if (typeof value === "string") {
        console.log("String value:", value.toUpperCase());
    } else {
        console.log("Number value:", value.toFixed(2));
    }
}

// --------------------
// Usage Example
// --------------------
(async () => {
    const user: User = {
        id: 1,
        name: username,
        role: isAdmin ? Role.Admin : Role.User,
    };

    log(greet(user));

    const service = new UserService([user]);
    service.addUser({ id: 2, name: "Dana", role: Role.Admin });

    console.log("Admins:", service.findByRole(Role.Admin));

    const fetchedUser = await fetchUser(3);
    const response = wrapResponse(fetchedUser);

    console.log("API response:", response);

    printValue("hello");
    printValue(123.456);

    console.log("2 + 3 =", add(2, 3));
})();
