import {
  Column,
  CreateDateColumn,
  Entity,
  PrimaryGeneratedColumn,
} from 'typeorm';

@Entity('areas')
export class Area {
  @PrimaryGeneratedColumn()
  id!: number;

  @Column({ length: 100 })
  name!: string;

  @Column()
  action!: string;

  @Column()
  reaction!: string;

  @Column({ default: true })
  enabled!: boolean;

  @Column({ type: 'timestamptz', nullable: true })
  lastTriggeredAt!: Date | null;

  @CreateDateColumn({ type: 'timestamptz' })
  createdAt!: Date;
}
