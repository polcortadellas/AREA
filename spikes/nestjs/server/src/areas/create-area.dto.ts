import { ApiProperty } from '@nestjs/swagger';
import { IsNotEmpty, IsString, MaxLength } from 'class-validator';

export class CreateAreaDto {
  @ApiProperty({ example: 'New GitHub issue -> Discord message' })
  @IsString()
  @IsNotEmpty()
  @MaxLength(100)
  name!: string;

  @ApiProperty({ example: 'github.new_issue' })
  @IsString()
  @IsNotEmpty()
  action!: string;

  @ApiProperty({ example: 'discord.send_message' })
  @IsString()
  @IsNotEmpty()
  reaction!: string;
}
