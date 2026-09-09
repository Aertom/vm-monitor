import type { VMFamily } from '../types/vm';

const colors: Record<VMFamily, string> = {
  sm: '#2563eb',
  cm: '#16a34a',
  ws: '#d97706',
  oa: '#9333ea',
  unknown: '#6b7280',
};

export function FamilyBadge({ family }: { family: VMFamily }) {
  return (
    <span
      style={{
        backgroundColor: colors[family],
        color: 'white',
        padding: '2px 8px',
        borderRadius: 4,
        fontSize: 12,
        textTransform: 'uppercase',
      }}
    >
      {family}
    </span>
  );
}
