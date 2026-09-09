import { useGroups } from './hooks/useGroups';
import { GroupTable } from './components/GroupTable';

export default function App() {
  const { data: groups, isLoading, error } = useGroups();

  if (isLoading) return <p>Chargement...</p>;
  if (error) return <p>Erreur de chargement des groupes.</p>;

  return (
    <div style={{ padding: 24, fontFamily: 'sans-serif' }}>
      <h1>VM Monitor</h1>
      {groups?.map((g) => (
        <GroupTable key={g.id} group={g} />
      ))}
    </div>
  );
}
