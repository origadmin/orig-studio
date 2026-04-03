/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import {useEffect, useState} from "react";
import {mediaApi, type EncodeProfile} from "../../lib/api/media";
import {Button} from "../../components/ui/button";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "../../components/ui/table";
import {Card, CardContent, CardHeader, CardTitle} from "../../components/ui/card";
import {Badge} from "../../components/ui/badge";
import {PlusCircle, Edit, Trash2} from "lucide-react";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
    DialogTrigger
} from "../../components/ui/dialog";
import {Input} from "../../components/ui/input";
import {Label} from "../../components/ui/label";

export default function TranscodingProfiles() {
    const [profiles, setProfiles] = useState<EncodeProfile[]>([]);
    const [loading, setLoading] = useState(true);
    const [editingProfile, setEditingProfile] = useState<Partial<EncodeProfile> | null>(null);
    const [isDialogOpen, setIsDialogOpen] = useState(false);

    const fetchProfiles = async () => {
        try {
            setLoading(true);
            const response = await mediaApi.listProfiles();
            // In our request.ts, api.get returns response.data directly
            setProfiles(response.profiles || []);
        } catch (error) {
            console.error("Failed to fetch profiles:", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchProfiles();
    }, []);

    const handleSave = async () => {
        if (!editingProfile) return;
        try {
            if (editingProfile.id) {
                await mediaApi.updateProfile(editingProfile.id, editingProfile);
            } else {
                await mediaApi.createProfile(editingProfile);
            }
            setIsDialogOpen(false);
            fetchProfiles();
        } catch (error) {
            console.error("Failed to save profile:", error);
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm("Are you sure you want to delete this profile?")) return;
        try {
            await mediaApi.deleteProfile(id);
            fetchProfiles();
        } catch (error) {
            console.error("Failed to delete profile:", error);
        }
    };

    const handleActivateRecommended = async () => {
        if (!confirm("This will activate h264 and h265 profiles for 240p, 360p, 480p, 720p, and 1080p resolutions. Continue?")) return;
        try {
            // Activate multiple profiles at once
            const profilesToActivate = [
                'h264-240', 'h264-360', 'h264-480', 'h264-720', 'h264-1080',
                'h265-240', 'h265-360', 'h265-480', 'h265-720', 'h265-1080'
            ];

            for (const name of profilesToActivate) {
                const profile = profiles.find(p => p.name === name);
                if (profile) {
                    await mediaApi.updateProfile(profile.id, {...profile, is_active: true});
                }
            }

            alert("Successfully activated recommended encoding profiles!");
            fetchProfiles();
        } catch (error) {
            console.error("Failed to activate profiles:", error);
            alert("Failed to activate profiles. Please activate them manually.");
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <h2 className="text-3xl font-bold tracking-tight">Encoding Profiles</h2>
                <div className="flex gap-2">
                    <Button variant="outline" onClick={handleActivateRecommended}>
                        Activate Recommended (10 profiles)
                    </Button>
                    <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                        <DialogTrigger asChild>
                            <Button onClick={() => setEditingProfile({is_active: true})}>
                                <PlusCircle className="mr-2 h-4 w-4"/>
                                Add Profile
                            </Button>
                        </DialogTrigger>
                        <DialogContent>
                            <DialogHeader>
                                <DialogTitle>{editingProfile?.id ? "Edit Profile" : "Add Profile"}</DialogTitle>
                            </DialogHeader>
                            <div className="grid gap-4 py-4">
                                <div className="grid grid-cols-4 items-center gap-4">
                                    <Label htmlFor="name" className="text-right">Name</Label>
                                    <Input id="name" value={editingProfile?.name || ""}
                                           onChange={(e) => setEditingProfile({
                                               ...editingProfile,
                                               name: e.target.value
                                           })} className="col-span-3"/>
                                </div>
                                <div className="grid grid-cols-4 items-center gap-4">
                                    <Label htmlFor="res" className="text-right">Resolution</Label>
                                    <Input id="res" placeholder="1920x1080" value={editingProfile?.resolution || ""}
                                           onChange={(e) => setEditingProfile({
                                               ...editingProfile,
                                               resolution: e.target.value
                                           })} className="col-span-3"/>
                                </div>
                                <div className="grid grid-cols-4 items-center gap-4">
                                    <Label htmlFor="vcodec" className="text-right">Video Codec</Label>
                                    <Input id="vcodec" placeholder="libx264" value={editingProfile?.video_codec || ""}
                                           onChange={(e) => setEditingProfile({
                                               ...editingProfile,
                                               video_codec: e.target.value
                                           })} className="col-span-3"/>
                                </div>
                                <div className="grid grid-cols-4 items-center gap-4">
                                    <Label htmlFor="vbitrate" className="text-right">Video Bitrate</Label>
                                    <Input id="vbitrate" placeholder="5000k" value={editingProfile?.video_bitrate || ""}
                                           onChange={(e) => setEditingProfile({
                                               ...editingProfile,
                                               video_bitrate: e.target.value
                                           })} className="col-span-3"/>
                                </div>
                            </div>
                            <DialogFooter>
                                <Button variant="outline" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
                                <Button onClick={handleSave}>Save</Button>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </div>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Management</CardTitle>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Name</TableHead>
                                <TableHead>Resolution</TableHead>
                                <TableHead>Video Codec</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow><TableCell colSpan={5}
                                                     className="text-center">Loading...</TableCell></TableRow>
                            ) : profiles.length === 0 ? (
                                <TableRow><TableCell colSpan={5} className="text-center">No profiles found.</TableCell></TableRow>
                            ) : (
                                profiles.map((p) => (
                                    <TableRow key={p.id}>
                                        <TableCell className="font-medium">{p.name}</TableCell>
                                        <TableCell>{p.resolution}</TableCell>
                                        <TableCell>{p.video_codec} ({p.video_bitrate})</TableCell>
                                        <TableCell>
                                            <Badge variant={p.is_active ? "default" : "secondary"}>
                                                {p.is_active ? "Active" : "Inactive"}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="text-right space-x-2">
                                            <Button variant="ghost" size="icon" onClick={() => {
                                                setEditingProfile(p);
                                                setIsDialogOpen(true);
                                            }}>
                                                <Edit className="h-4 w-4"/>
                                            </Button>
                                            <Button variant="ghost" size="icon" className="text-destructive"
                                                    onClick={() => handleDelete(p.id)}>
                                                <Trash2 className="h-4 w-4"/>
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}
