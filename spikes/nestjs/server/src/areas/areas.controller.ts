import { Body, Controller, Get, Post } from '@nestjs/common';
import { ApiCreatedResponse, ApiTags } from '@nestjs/swagger';
import { AreasService } from './areas.service';
import { CreateAreaDto } from './create-area.dto';

@ApiTags('areas')
@Controller('areas')
export class AreasController {
  constructor(private readonly areas: AreasService) {}

  @Get()
  findAll() {
    return this.areas.findAll();
  }

  @Post()
  @ApiCreatedResponse({ description: 'AREA created' })
  create(@Body() dto: CreateAreaDto) {
    return this.areas.create(dto);
  }
}
