import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Area } from './area.entity';
import { CreateAreaDto } from './create-area.dto';

@Injectable()
export class AreasService {
  constructor(
    @InjectRepository(Area) private readonly repo: Repository<Area>,
  ) {}

  findAll(): Promise<Area[]> {
    return this.repo.find({ order: { id: 'ASC' } });
  }

  findEnabled(): Promise<Area[]> {
    return this.repo.find({ where: { enabled: true } });
  }

  create(dto: CreateAreaDto): Promise<Area> {
    return this.repo.save(this.repo.create(dto));
  }

  async markTriggered(id: number): Promise<void> {
    await this.repo.update(id, { lastTriggeredAt: new Date() });
  }
}
