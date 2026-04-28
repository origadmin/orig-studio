import * as React from "react"
import {cn} from "@/lib/utils"

export interface SpinnerProps extends React.HTMLAttributes<HTMLDivElement> {
    size?: "sm" | "default" | "lg"
}

const spinnerSizeMap = {
    sm: "w-4 h-4 border-2",
    default: "w-8 h-8 border-[3px]",
    lg: "w-12 h-12 border-4",
}

const Spinner = React.forwardRef<HTMLDivElement, SpinnerProps>(
    ({size = "default", className, ...props}, ref) => {
        return (
            <div
                role="status"
                aria-label="加载中"
                ref={ref}
                className={cn(
                    "animate-spin rounded-full border-primary border-t-transparent",
                    spinnerSizeMap[size],
                    className
                )}
                {...props}
            />
        )
    }
)
Spinner.displayName = "Spinner"

export {Spinner}
