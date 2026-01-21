// Sample.tsx
// A practical TSX + React + TypeScript example 🧩

import React, { useEffect, useState } from "react";

// --------------------
// Types & Interfaces
// --------------------
type Role = "USER" | "ADMIN";

interface User {
    id: number;
    name: string;
    role: Role;
}

// Props for a component
interface UserCardProps {
    user: User;
    onSelect: (id: number) => void;
}

// --------------------
// Reusable Component
// --------------------
const UserCard: React.FC<UserCardProps> = ({ user, onSelect }) => {
    return (
        <div style={{ border: "1px solid #ccc", padding: 12, marginBottom: 8 }}>
            <h3>{user.name}</h3>
            <p>Role: {user.role}</p>
            <button onClick={() => onSelect(user.id)}>Select</button>
        </div>
    );
};

// --------------------
// Main Component
// --------------------
const Sample: React.FC = () => {
    // Typed state
    const [users, setUsers] = useState<User[]>([]);
    const [selectedUserId, setSelectedUserId] = useState<number | null>(null);
    const [loading, setLoading] = useState<boolean>(true);

    // --------------------
    // Side Effects
    // --------------------
    useEffect(() => {
        // Fake async fetch
        const loadUsers = async () => {
            setLoading(true);

            await new Promise(resolve => setTimeout(resolve, 500));

            setUsers([
                { id: 1, name: "Alice", role: "USER" },
                { id: 2, name: "Bob", role: "ADMIN" },
            ]);

            setLoading(false);
        };

        loadUsers();
    }, []);

    // --------------------
    // Event Handlers
    // --------------------
    const handleSelectUser = (id: number) => {
        setSelectedUserId(id);
    };

    // --------------------
    // Derived Values
    // --------------------
    const selectedUser = users.find(u => u.id === selectedUserId);

    // --------------------
    // Conditional Rendering
    // --------------------
    if (loading) {
        return <div>Loading users...</div>;
    }

    return (
        <div style={{ maxWidth: 400, margin: "0 auto" }}>
            <h2>User List</h2>

            {users.map(user => (
                <UserCard
                    key={user.id}
                    user={user}
                    onSelect={handleSelectUser}
                />
            ))}

            {selectedUser && (
                <div style={{ marginTop: 16 }}>
                    <strong>Selected User:</strong>
                    <div>{selectedUser.name}</div>
                </div>
            )}
        </div>
    );
};

export default Sample;
