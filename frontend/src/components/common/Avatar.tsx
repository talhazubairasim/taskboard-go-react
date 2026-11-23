interface AvatarProps {
  name: string;
  src?: string | null;
  size?: 'sm' | 'md' | 'lg';
}

const sizes = { sm: 'h-6 w-6 text-xs', md: 'h-8 w-8 text-sm', lg: 'h-16 w-16 text-2xl' };

export const Avatar = ({ name, src, size = 'md' }: AvatarProps) =>
  src ? (
    <img src={src} alt={name} className={`${sizes[size]} rounded-full object-cover`} />
  ) : (
    <span
      aria-label={name}
      className={`${sizes[size]} inline-flex items-center justify-center rounded-full bg-indigo-600 font-medium text-white`}
    >
      {name.trim().charAt(0).toUpperCase() || '?'}
    </span>
  );
