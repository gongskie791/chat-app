import { createRoom, getRooms } from "@/api/rooms";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useAuth } from "@/context/AuthContext";
import type { Room } from "@/types";
import { Hash, LogOut, Plus } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

export default function RoomsPage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const [rooms, setRooms] = useState<Room[]>([]);
  const [newRoomName, setNewRoomName] = useState("");
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);

  useEffect(() => {
    getRooms()
      .then((res) => setRooms(res.data ?? []))
      .catch(() => setError("Failed to load rooms"));
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!newRoomName.trim()) return;
    setCreating(true);
    setError("");
    try {
      const res = await createRoom(newRoomName.trim());
      if (res.data) {
        setRooms((prev) => [res.data!, ...prev]);
        setNewRoomName("");
        setShowForm(false);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create room");
    } finally {
      setCreating(false);
    }
  }

  async function handleLogout() {
    await logout();
    navigate("/login");
  }

  return (
    <div className="flex h-screen flex-col bg-background">
      {/* Header */}
      <header className="flex items-center justify-between border-b border-border px-6 py-3">
        <div className="flex items-center gap-2">
          <Hash className="h-5 w-5 text-muted-foreground" />
          <span className="font-semibold text-foreground">Chat App</span>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-sm text-muted-foreground">
            {user?.username}
          </span>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleLogout}
            className="gap-1.5 text-muted-foreground hover:text-foreground"
          >
            <LogOut className="h-4 w-4" />
            Sign out
          </Button>
        </div>
      </header>

      {/* Body */}
      <main className="mx-auto w-full max-w-2xl flex-1 overflow-y-auto px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-foreground">Rooms</h2>
            <p className="text-sm text-muted-foreground">
              {rooms.length} room{rooms.length !== 1 ? "s" : ""} available
            </p>
          </div>
          <Button size="sm" onClick={() => setShowForm((v) => !v)}>
            <Plus className="h-4 w-4" />
            New room
          </Button>
        </div>

        {/* Create form */}
        {showForm && (
          <form
            onSubmit={handleCreate}
            className="mb-6 flex items-center gap-2"
          >
            <Input
              placeholder="Room name…"
              value={newRoomName}
              onChange={(e) => setNewRoomName(e.target.value)}
              autoFocus
              minLength={2}
              maxLength={100}
            />
            <Button type="submit" disabled={creating} size="sm">
              {creating ? "Creating…" : "Create"}
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setShowForm(false)}
            >
              Cancel
            </Button>
          </form>
        )}

        {error && (
          <p className="mb-4 rounded-md bg-destructive/10 px-3 py-2 text-xs text-destructive">
            {error}
          </p>
        )}

        {/* Room list */}
        <div className="space-y-2">
          {rooms.length === 0 ? (
            <div className="flex flex-col items-center gap-2 py-16 text-center">
              <Hash className="h-8 w-8 text-muted-foreground/40" />
              <p className="text-sm text-muted-foreground">
                No rooms yet. Create the first one.
              </p>
            </div>
          ) : (
            rooms.map((room) => <RoomCard key={room.id} room={room} />)
          )}
        </div>
      </main>
    </div>
  );
}

function RoomCard({ room }: { room: Room }) {
  const navigate = useNavigate();

  return (
    <Card
      className="cursor-pointer transition-colors hover:bg-accent/50"
      onClick={() => navigate(`/rooms/${room.id}`)}
    >
      <CardContent className="flex items-center justify-between px-4 py-3">
        <div className="flex items-center gap-3">
          <Hash className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span className="font-medium text-foreground">{room.name}</span>
        </div>
        <Badge variant="secondary" className="text-xs text-muted-foreground">
          {new Date(room.created_at).toLocaleDateString()}
        </Badge>
      </CardContent>
    </Card>
  );
}
