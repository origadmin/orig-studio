/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import {useEffect, useState} from "react";
import {mediaApi, type EncodeProfile} from "../../lib/api/media";
import {Button} from "../../components/ui/button";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "../../components/ui/table";
import {Card, CardContent, CardHeader, CardTitle} from "../../components/ui/card";
import {Badge} from "../../components/ui/badge";
import {PlusCircle, Edit, Trash2, CheckCircle, XCircle} from "lucide-react";
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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "../../components/ui/select";
import {Checkbox} from "../../components/ui/checkbox";

export default function TranscodingProfiles() {
    const [profiles, setProfiles] = useState<EncodeProfile[]>([]);
    const [loading, setLoading] = useState(true);
    const [editingProfile, setEditingProfile] = useState<Partial<EncodeProfile> | null>(null);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [selectedCodec, setSelectedCodec] = useState('h264');
    const [selectedResolutions, setSelectedResolutions] = useState<string[]>(['240', '360', '480', '720', '1080']);

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
        if (selectedResolutions.length === 0) {
            alert("Please select at least one resolution");
            return;
        }

        if (!confirm(`This will activate ${selectedCodec.toUpperCase()} profiles for the selected resolutions. Continue?`)) return;
        try {
            // Activate selected profiles
            const profilesToActivate = selectedResolutions.map(res => `${selectedCodec}-${res}`);

            let activatedCount = 0;
            for (const name of profilesToActivate) {
                const profile = profiles.find(p => p.name === name);
                if (profile) {
                    await mediaApi.updateProfile(profile.id, {...profile, is_active: true});
                    activatedCount++;
                }
            }

            alert(`Successfully activated ${activatedCount} encoding profiles!`);
            fetchProfiles();
        } catch (error) {
            console.error("Failed to activate profiles:", error);
            alert("Failed to activate profiles. Please activate them manually.");
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                <h2 className="text-3xl font-bold tracking-tight">Encoding Profiles</h2>
                <div className="flex flex-col sm:flex-row gap-4 w-full sm:w-auto">
                    <div className="flex gap-2">
                        <Select value={selectedCodec} onValueChange={setSelectedCodec}>
                            <SelectTrigger className="w-[120px]">
                                <SelectValue placeholder="Codec"/>
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="h264">H.264</SelectItem>
                                <SelectItem value="h265">H.265</SelectItem>
                            </SelectContent>
                        </Select>
                        <div className="flex flex-wrap gap-2">
                            {['240', '360', '480', '720', '1080'].map((res) => (
                                <div key={res} className="flex items-center space-x-1">
                                    <Checkbox
                                        id={`res-${res}`}
                                        checked={selectedResolutions.includes(res)}
                                        onCheckedChange={(checked) => {
                                            if (checked) {
                                                setSelectedResolutions([...selectedResolutions, res]);
                                            } else {
                                                setSelectedResolutions(selectedResolutions.filter(r => r !== res));
                                            }
                                        }}
                                    />
                                    <label htmlFor={`res-${res}`}
                                           className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                        {res}p
                                    </label>
                                </div>
                            ))}
                        </div>
                    </div>
                    <div className="flex gap-2">
                        <Button variant="outline" onClick={handleActivateRecommended}>
                            Activate Selected
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
                                        <Input id="vcodec" placeholder="libx264"
                                               value={editingProfile?.video_codec || ""}
                                               onChange={(e) => setEditingProfile({
                                                   ...editingProfile,
                                                   video_codec: e.target.value
                                               })} className="col-span-3"/>
                                    </div>
                                    <div className="grid grid-cols-4 items-center gap-4">
                                        <Label htmlFor="vbitrate" className="text-right">Video Bitrate</Label>
                                        <Input id="vbitrate" placeholder="5000k"
                                               value={editingProfile?.video_bitrate || ""}
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
