/*
 * Copyright 2009 Mukunda Johnson (www.mukunda.com)
 * 
 * This file is part of SNESMOD.
 * 
 * SNESMOD is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * 
 * SNESMOD is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * 
 * You should have received a copy of the GNU General Public License
 * along with SNESMOD.  If not, see <http://www.gnu.org/licenses/>.
 */
 
// a class to read a binary file with stdio

#ifndef __IO_H
#define __IO_H

#include <stdio.h>
#include "basetypes.h"

namespace IO {

	typedef enum {
		MODE_READ,
		MODE_WRITE
	} FileAccessMode;

	class File {

	private:
		FILE	*file;
		bool	isOpen;
		FileAccessMode mode;

	public:
		File();
		File( const char *, FileAccessMode );
		~File();
		bool Open( const char *, FileAccessMode );
		void Close();
		
		u8 Read8();
		u16 Read16();
		u32 Read32();

		void Write8( u8 );
		void Write16( u16 );
		void Write32( u32 );

		// write non-terminated ascii
		void WriteAscii( const char * );

		// write fixed length string (zero padded)
		void WriteAsciiF( const char *, int );

		void ZeroFill( int );

		void WriteAlign( u32 boundary );
		
		void Skip( s32 amount );
		u32 Tell();
		void Seek( u32 offset );

		bool IsOpen();

	};

	bool FileExists();
	u32  FileSize( const char *filename );

};

#endif
