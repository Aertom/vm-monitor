import type { Group } from '../types/vm';
import { FamilyBadge } from './FamilyBadge';
import { useCheckout } from '../hooks/useCheckout';
import { useState } from 'react';

export function GroupTable({ group }: { group: Group }) {
  const { checkout, checkin } = useCheckout(group.id);
  const [userName, setUserName] = useState('');

  return (
    <div style={{ marginBottom: 32 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
        <h3 style={{ margin: 0 }}>Groupe: {group.id}</h3>
        {group.inUseBy ? (
          <>
            <span style={{ color: '#d97706' }}>
              En cours d'utilisation par <strong>{group.inUseBy}</strong>
            </span>
            <button onClick={() => checkin()}>Libérer</button>
          </>
        ) : (
          <>
            <input
              placeholder="Votre nom"
              value={userName}
              onChange={(e) => setUserName(e.target.value)}
              style={{ padding: 4 }}
            />
            <button disabled={!userName} onClick={() => checkout(userName)}>
              Utiliser ce groupe
            </button>
          </>
        )}
      </div>

      <table border={1} cellPadding={6} style={{ borderCollapse: 'collapse', width: '100%' }}>
        <thead>
          <tr>
            <th>Famille</th>
            <th>Hostname</th>
            <th>IP</th>
            <th>Hyperviseur</th>
            <th>Versions applis</th>
            <th>Dernière vérif.</th>
          </tr>
        </thead>
        <tbody>
          {group.vms.map((vm) => (
            <tr key={vm.id}>
              <td><FamilyBadge family={vm.family} /></td>
              <td>{vm.hostname}</td>
              <td>{vm.ip}</td>
              <td>{vm.hypervisor}</td>
              <td>
                {vm.appVersions
                  ? Object.entries(vm.appVersions)
                      .map(([app, v]) => `${app}: ${v}`)
                      .join(', ')
                  : vm.error ?? '—'}
              </td>
              <td>{new Date(vm.lastChecked).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
