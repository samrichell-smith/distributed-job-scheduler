declare module 'react-window' {
  import * as React from 'react';

  export interface ListChildComponentProps {
    index: number;
    style: React.CSSProperties;
    key?: React.Key;
  }

  export interface FixedSizeListProps {
    height: number;
    itemCount: number;
    itemSize: number;
    width: number | string;
    children: React.ComponentType<ListChildComponentProps> | ((props: ListChildComponentProps) => React.ReactElement) ;
  }

  export class FixedSizeList extends React.Component<FixedSizeListProps> {}

  export default FixedSizeList;
}
