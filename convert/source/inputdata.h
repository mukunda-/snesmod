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
 
#ifndef __INPUTDATA_H
#define __INPUTDATA_H

// param parser

#include <stdlib.h>
#include <string>
#include <vector>
#include "basetypes.h"

namespace ConversionInput {
/*
	class EffectData {
	public:
		EffectData( const TiXmlElement * );
		std::string filename;
		std::string id;
		
		int force_filter;
		
	};
*/
	/*
	class SampleData {

	public:
		int index;
		int force_filter;
		int force_loop_filter;
		std::string id;
	
		SampleData( const TiXmlElement * );
		SampleData( const EffectData & );
	};
*/
	/*
	class ModuleData {

	private:
		u8 ConvertBitString( const char * );
		u8 TranslatePercentage( int );
		
	public:

		ModuleData( const TiXmlElement *source );
		~ModuleData();
		std::string filename;
		std::string id;
		
		u8	EDL;
		u8	EFB;
		u8	EVL;
		u8	EVR;
		u8	EON;
		u8	COEF[8];
		
		std::vector<SampleData*> samples;
	};
*/
	/*
	typedef struct {

		std::vector<const char *> files;
		bool hirom;
		std::string output;
	} SoundbankInfo;

	class SoundbankData {

	public:
		
		SoundbankData( const SoundbankInfo &info );
		
		std::string output;
		bool hirom;
		
		std::vector<ModuleData*> modules;
		std::vector<EffectData*> effects;
	};
*/
	class OperationData {

	private:
		

	public:

		~OperationData();

		OperationData( int argc, char *argv[] );

		std::string output;
		bool hirom;
		std::vector<const char *> files;
		bool spc_mode;
		bool verbose_mode;
		bool show_help;

	};
}

#endif
