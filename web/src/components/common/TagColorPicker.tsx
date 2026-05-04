import React, {useState} from 'react';
import {Popover, PopoverContent, PopoverTrigger} from '@/components/ui/popover';
import {Input} from '@/components/ui/input';
import {Button} from '@/components/ui/button';
import {cn} from '@/lib/utils';
import {TAG_COLOR_PALETTE, isValidHexColor} from '@/lib/utils/tag-color';

interface TagColorPickerProps {
    value?: string;
    onChange?: (color: string) => void;
}

const TagColorPicker: React.FC<TagColorPickerProps> = ({value, onChange}) => {
    const [open, setOpen] = useState(false);
    const [inputValue, setInputValue] = useState(value || '');

    const handleSelect = (color: string) => {
        setInputValue(color);
        onChange?.(color);
        setOpen(false);
    };

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const v = e.target.value;
        setInputValue(v);
        if (isValidHexColor(v)) {
            onChange?.(v);
        }
    };

    const handleInputBlur = () => {
        if (inputValue && !isValidHexColor(inputValue)) {
            setInputValue(value || '');
        }
    };

    const handleClear = () => {
        setInputValue('');
        onChange?.('');
        setOpen(false);
    };

    const displayColor = value && isValidHexColor(value) ? value : '';

    return (
        <Popover open={open} onOpenChange={setOpen}>
            <PopoverTrigger asChild>
                <Button
                    variant="outline"
                    role="combobox"
                    aria-expanded={open}
                    className="w-full justify-start gap-2 h-9 font-normal"
                >
                    <span
                        className={cn(
                            'h-4 w-4 rounded-full border border-gray-300 dark:border-gray-600 shrink-0',
                            !displayColor && 'bg-gray-200 dark:bg-gray-700'
                        )}
                        style={displayColor ? {backgroundColor: displayColor} : undefined}
                    />
                    <span className="truncate text-muted-foreground text-xs">
                        {displayColor || 'Auto (hash)'}
                    </span>
                </Button>
            </PopoverTrigger>
            <PopoverContent className="w-56 p-3" align="start">
                <div className="space-y-3">
                    <div className="grid grid-cols-6 gap-1.5">
                        {TAG_COLOR_PALETTE.map((color) => (
                            <button
                                key={color}
                                type="button"
                                className={cn(
                                    'h-6 w-6 rounded-md border-2 transition-colors hover:scale-110',
                                    value === color
                                        ? 'border-gray-900 dark:border-gray-100'
                                        : 'border-transparent'
                                )}
                                style={{backgroundColor: color}}
                                onClick={() => handleSelect(color)}
                            />
                        ))}
                    </div>
                    <div className="flex gap-2">
                        <Input
                            placeholder="#000000"
                            value={inputValue}
                            onChange={handleInputChange}
                            onBlur={handleInputBlur}
                            className="h-8 text-xs font-mono"
                            maxLength={7}
                        />
                        <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            className="h-8 px-2 text-xs shrink-0"
                            onClick={handleClear}
                        >
                            Clear
                        </Button>
                    </div>
                </div>
            </PopoverContent>
        </Popover>
    );
};

export default TagColorPicker;
