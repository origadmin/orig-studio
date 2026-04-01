import React, {useState} from 'react';
import {mediaApi} from '../lib/api/media';
import {Button} from './ui/button';
import {Card, CardHeader, CardTitle, CardContent, CardFooter} from './ui/card';

interface MediaUploadProps {
    onSuccess?: (media: any) => void;
    userId: string;
}

export const MediaUpload: React.FC<MediaUploadProps> = ({onSuccess, userId}) => {
    const [file, setFile] = useState<File | null>(null);
    const [title, setTitle] = useState('');
    const [description, setDescription] = useState('');
    const [uploading, setUploading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files[0]) {
            const selectedFile = e.target.files[0];
            setFile(selectedFile);
            if (!title) setTitle(selectedFile.name);
        }
    };

    const handleUpload = async () => {
        if (!file) return;

        setUploading(true);
        setError(null);

        try {
            const media = await mediaApi.upload(file, {
                title,
                description,
                user_id: userId
            });
            setFile(null);
            setTitle('');
            setDescription('');
            if (onSuccess) onSuccess(media);
            alert('Upload successful!');
        } catch (err: any) {
            setError(err.message || 'Upload failed');
        } finally {
            setUploading(false);
        }
    };

    return (
        <Card className="w-full max-w-md mx-auto">
            <CardHeader>
                <CardTitle>Upload Media</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
                <div>
                    <label className="block text-sm font-medium mb-1">File</label>
                    <input
                        type="file"
                        onChange={handleFileChange}
                        accept="video/*,image/*"
                        className="w-full border rounded p-2 text-sm"
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium mb-1">Title</label>
                    <input
                        type="text"
                        value={title}
                        onChange={(e) => setTitle(e.target.value)}
                        placeholder="Enter title"
                        className="w-full border rounded p-2 text-sm"
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium mb-1">Description</label>
                    <textarea
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        placeholder="Enter description"
                        className="w-full border rounded p-2 text-sm h-24"
                    />
                </div>
                {error && <p className="text-red-500 text-sm">{error}</p>}
            </CardContent>
            <CardFooter>
                <Button
                    onClick={handleUpload}
                    disabled={!file || uploading}
                    className="w-full"
                >
                    {uploading ? 'Uploading...' : 'Upload'}
                </Button>
            </CardFooter>
        </Card>
    );
};
