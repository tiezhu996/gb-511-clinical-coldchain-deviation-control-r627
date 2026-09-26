import DeviceThermostatOutlinedIcon from '@mui/icons-material/DeviceThermostatOutlined';

export function TemperatureBadge({ value, minimum, maximum }: { value: number; minimum?: number; maximum?: number }) {
  const outside = (minimum !== undefined && value < minimum) || (maximum !== undefined && value > maximum);
  return <span className={`temperature-badge ${outside ? 'temperature-badge--alarm' : 'temperature-badge--normal'}`} title={outside ? '温度超出规则窗口' : '温度位于可接受范围'}><DeviceThermostatOutlinedIcon />{value.toFixed(1)} C</span>;
}
