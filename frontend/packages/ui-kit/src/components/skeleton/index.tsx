import { cn } from "#lib/utils";

function SkeletonRoot({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="skeleton"
      className={cn("animate-pulse rounded-md bg-muted", className)}
      {...props}
    />
  );
}

export const Skeleton = SkeletonRoot;
