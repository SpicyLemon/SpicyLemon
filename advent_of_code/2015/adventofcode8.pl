#!/usr/bin/perl
use strict;
use warnings;
#use Benchmark ':hireswallclock';
#use Carp;
use Data::Dumper;  $Data::Dumper::Sortkeys = 1;
use DateTime;
#use Date::Parse; #imports str2time that takes in a string, and outputs an epoch.
#use Digest::MD5 qw(md5_hex);
#use Fcntl;
#use File::Copy;   #imports copy and move
#use HTML::Entities;
#use List::Util 'shuffle';
#use Math::BigInt;
#use Math::Round qw(nearest);
#use SDBM_File;
#use Storable qw(dclone);
#use Time::HiRes qw(time usleep);  #this forces the time() function to return nanosecond information
#use XML::Simple;

#use easyopen qw(openr openread openw openwrite opena openappend slurpFile opend slurpDir);
#use fake_list_data qw(d l L c w date firstname lastname fullname business address city state zip zip4 location amount purpose word rib ri r Time dateAndTime);
#use list_helpers qw(inOneListButNotTheOther inOneStrButNotTheOther strToList strToHash strToListOfHashes isIn);
#use logger qw(printlog dielog dieMessage logfile printlogWithBorder dielogWithBorder addBorderTo);
#use printers; #qw(listOfHashRefs csv hash indent table listForSql breakMultiLine prettyList);
      #$output = printers::table($data, $col_order, $delimeter);
use Prompt;   #imports the prompt($string, $default, $required) function
#use stats qw(getStats countMinMax mean median modes);

#use lib qw(.);


mainProgram();
print "Done\n";
exit(0);

sub mainProgram {
   print "Test:\n";
   mainProgramTest();
   print "part 1:\n";
   mainProgram1();
   print "part 2:\n";
   mainProgram2();
}

sub mainProgramTest {
   my @tests = (
      { input => '""', expected1 => 0, expected2 => 2 },
      { input => '"abc"', expected1 => 3, expected2 => 5 },
      { input => '"aaa\"aaa"', expected1 => 7, expected2 => 10 },
      { input => '"\x27"', expected1 => 1, expected2 => 6 },
      { input => '"\xa8br\x8bjr\""', expected1 => 7, expected2 => 16 },
      { input => '"daz\\\\zyyxddpwk"', expected1 => 13, expected2 => 16 },
   );
   my $all_good = 1;
   foreach my $test (@tests) {
      my $actual1 = countMemChars($test->{input});
      my $actual2 = length($test->{input});
      if ($actual1 != $test->{expected1}) {
         $all_good = 0;
         print "Test 1 failed: ".$test->{input}.", exptected: ".$test->{expected1}.", actual: ".$actual1."\n";
      }
      if ($actual2 != $test->{expected2}) {
         $all_good = 0;
         print "Test 2 failed: ".$test->{input}.", exptected: ".$test->{expected2}.", actual: ".$actual2."\n";
      }
   }
   if ($all_good) {
      print "All tests passed.\n";
   }
}


sub mainProgram1 {
   my $coded_length_total = 0;
   my $actual_length_total = 0;
   foreach my $str (split(/\n/, getStr())) {
      my $coded_length = length($str);
      my $actual_length = countMemChars($str);
      $coded_length_total += $coded_length;
      $actual_length_total += $actual_length;
      #prompt($str.' - '.$coded_length.', '.$actual_length , "", 0)
   }
   my $diff = $coded_length_total - $actual_length_total;
   print "coded length : $coded_length_total\n";
   print "actual length: $actual_length_total\n";
   print "difference   : $diff\n";
}


sub mainProgram2 {
   my $coded_length_total = 0;
   my $recoded_length_total = 0;
   foreach my $str (split(/\n/, getStr())) {
      my $coded_length = length($str);
      my $recoded_length = countReEncodedChars($str);
      $coded_length_total += $coded_length;
      $recoded_length_total += $recoded_length;
      #prompt($str.' - '.$coded_length.', '.$recoded_length , "", 0)
   }
   my $diff = $recoded_length_total - $coded_length_total;
   print "recoded length: $recoded_length_total\n";
   print "coded length  : $coded_length_total\n";
   print "difference    : $diff\n";
}


sub countMemChars {
   my $input = shift;
   my $output = $input;
   $output =~ s/^"//;
   $output =~ s/"$//;
   $output =~ s/\\"/"/g;
   $output =~ s/\\\\/\\/g;
   $output =~ s/\\x[abcdefgABCDEFG0123456789]{2}/_/g;
   return length($output);
}



sub countReEncodedChars {
   my $input = shift;
   my $output = $input;
   $output =~ s/\\/\\\\/g;
   $output =~ s/"/\\"/g;
   $output = '"'.$output.'"';
   return length($output);
}

#Note, I changed // to //// in this input because perl was automatically change // to /, and I needed it still as //
sub getStr {
   return q^"\xa8br\x8bjr\""
"nq"
"zjrfcpbktjmrzgsz\xcaqsc\x03n\"huqab"
"daz\\\\zyyxddpwk"
"draes\xa2n\\\\g\x27ek\"lj\"\\\\viqych"
"nnx\\\\krnrfomdnt\x2flbl\xd2xpo\"cp\"k"
"kwdaapalq"
"u\"ptk"
"ckhorczuiudfjmmcc\\\\u\"wozqxibsfjma"
"ydctdrxat\"pd\"lwi\"bjesevfw\xe8"
"v\"\xa8rrzep\"\"r"
"nbydghkfvmq\\\\\xe0\"lfsrsvlsj\"i\x61liif"
"jsas\"u"
"odipikxlo"
"\"rnubsgwltqkbsu\"pcpcs"
"eitk\\\\f\\\\mhcqqoym\\\\ji"
"vnedc"
"\"lhcaurdqzyjyu"
"haxzsa\"zcn\"y\"foclgtjfcnv\"m\x68krc"
"\"eoeggg\"tmiydvcay\"vfavc"
"snqvyqoncwxcvwbdktoywch"
"rnfgjsyr\xd5wacy"
"ik\"hebrpvsts"
"txw"
"\x15pxtdkogd\"urbm\"gevhh\"nxr\x3erxtk"
"cetqtcy"
"inleep\\\\mgl"
"uflwbxvww\x2cxzezqnaply\"yh\"qlllzk"
"eepak\"xqtedzt"
"na\x61qzfieafvyrsnwkssznohjmc"
"yceaonylz\xc1\\\\jrlbbkzwsidfi"
"ybqafngkcqpbp"
"\xaft"
"yidjpaobqydso"
"ju\\\\ldxig\\\\lrdrhjcmm\x77rc"
"tylacqeslnwj\x48ds\"tjxa"
"efbfm"
"\\\\fxkgoprgdcjgyajykg\\\\dtbrz"
"eujvva"
"h\x7acwfpikme\\\\vwthyvrqdnx\""
"rbpbrxm\\\\\"\"\"voxx"
"ykiw\"tkb\\\\lforu\"rsf\\\\tf\"x\"rqti"
"e\\\\wh\x77aqeugiq\\\\ihhfqfuaij"
"g\"t\\\\o"
"nxzo\"hf\\\\xp"
"dxiaqfo\xea"
"kali\\\\zczhiqkqzybjj\"fgdjnik"
"zdkgrqmdv"
"bimxim\xb6lrwsaj\"ui\"a"
"\"rrznitibgx\\\\olpsjmjqzctxaubdifsq"
"zb\"khzixaacmhuzmlymoformipdzml"
"qfwi"
"hjwsxfpphttjy\"\"zixais\xbblgnqfto"
"puj\\\\qmyu\"nqgaqfthbwjokbmrpbhpi"
"cyxdpkh\\\\\""
"q"
"m"
"tbxdzzllarlo"
"gbtys"
"gytilk\\\\vlqxvcuutjunrqc"
"uugkvcuzan\\\\eyhb"
"yaxr\"genlbgw\"\\\\uc"
"nrgecjeip\\\\sjdvgqaqxwsqactopu"
"pu\"r\"txpyrkfny\\\\zmwfneyvwmnkkdipv"
"jm\xa3bhwvq"
"qxojmnml\"w\x9airr"
"xbzsuihs\x4dcedy\xaclrhgii\\\\\""
"drgjirusrekrwmvxllwdm"
"\x28hfxnfpycmpnkku\"csuf\xaarxlqyg\"x"
"\"zvz\\\\rmg\"\\\\sxxoifffyqfyn\"iq\"ps"
"\"z"
"zbwkmk\"sgzos\x93gtc\""
"bvm\x28aa\\\\\\\\\"pywuhaniox\\\\z\\\\hbp\xd7mold"
"aszgvsyna"
"qf\"vdwuss"
"lnohni\"qwiacjsjegstlbfq\\\\kyjhyd"
"c\\\\naawulxlqplnacvytspry\xf5ytxxqq"
"razwqmsqgbaaxcd\\\\f"
"radggyrjrg\"zx"
"\"pu\x11t\\\\ajcjuieinlkvya"
"veggiskh"
"eglfhjxiet\"kouqfskwsy\"hpthsldel"
"mv\xc1b\"f\\\\shrssnjwcpmurepdxdlcj"
"dlayjd\"suvzotgdtc"
"\xa9pvxeopn"
"lpplsaxy\"oiwaq"
"hqwh\\\\lusv"
"hykykwlx\"\xa5atkgh\\\\d\x63dff"
"vfktanpjy\"xxetc"
"dnhwkgjnsmsswfuelvihvjl\"jtf"
"x\"dwvzra\"nbbsewftehczgbvfzd\"rau"
"csfi\"mzejnjqkqupwadrgti\"von"
"xckf\xf7xsm\\\\pgvlpetjndpyblais\\\\z"
"yecy\x6fuj\x58bwpgeuiw\"mdu"
"fgb"
"c\\\\lx\x3efthet\xfdelgvwvpem"
"kgyrmarvfwjinlowt"
"yzte"
"vc\"z"
"sxevqfzmmdwsuu\""
"fxbaercmcy\xb6md"
"f"
"m\x44gqbcppho\\\\b"
"gtafr\x57m\x11jy\"\"erwmmpiwjkbckuw"
"ufdjt\"kssprzxqixzxmq\x58q"
"yzbyo\"lfdbyaxexyfbnyv\\\\\xe8xmre"
"u\x43ntr\\\\\\\\byyfjr\"iveujvnwsqbnpuvrta"
"us\xf6bai"
"c\\\\edh"
"tzckolphexfq\\\\\x23\xfbdqv\\\\\"m"
"yjafhbvhhj\x1b\"bplb"
"\"o"
"rubahvmp\""
"qmkukrnrmqumh"
"wdpxyvyidhwjf\\\\nabbijwhr\xc5bksvy\"p"
"u\"prlpg\""
"nsvcquyxbwilsxxemf\xd9leq"
"y\xcetxuafl"
"it"
"kwdlysf\\\\xjpelae"
"viwh\x58wpjjlnvryuti\x2chngrx\\\\nhtkui"
"vhn\x9ehre\xc3ncsqbozms\"nl"
"ytc\xa3mgeeogjcqavmmmd"
"xzlexlitseozoxtpzzutfq"
"cish\x07lmovj"
"ekbflwqzaiivdr\"pq\\\\azrfbntqwkn"
"uc\"xdbegmlmhksofzohavtrnxf"
"xfdnrdqdrcjzbe"
"ndg\"ckgrpisib\"rg\"p\\\\lmpfzlssnvk"
"witfjwpbyyzlop"
"zonlww\"emrbcsgdtrg\"rjzy\x64zqntlw"
"dvgb\"zn\\\\vrbzema\"ckmd"
"\\\\vdlmxhlvldk\"pmzazeip"
"\"\"r"
"rsntinv"
"iy"
"lr\x20efh"
"csgexlb\"zqdavlxxhtdbh\"\"\x0fkpvhiphm"
"ouwhp\"ogbft"
"cm\\\\ckltng\"dw\x8brf\xf0eppgckd"
"zmnlsgalhpkejsizfsbtnfliu\"nhc"
"pnrkaayqvwpdjbhcrbb\"yfeq\"aq"
"ozh\\\\hoxow\x2csrtr\\\\r\""
"bqxabj\"u\\\\s"
"cpsjti\"gy"
"aa\"p\\\\nki\\\\ajijkqev"
"q\"\"lfdentjgd\\\\"
"bmokvpoebutfki"
"pielvcbne\xf6efvzxn"
"kx"
"zlgmqagcrbhrwtwtmmg"
"aiyhmntcqjbpv\xb5hhswxbryoedvos"
"tdpaxrb"
"fu\"\x7dttkyvhrlwko"
"oirc\\\\\"cqlnqffjqt\\\\k"
"edxlia\\\\tcyby"
"jpeybgwfayerfrfbvfog\"ol"
"ysr"
"bzwzilgwfugjk"
"tlcc\x75nukvwjgftetjcs\xaecwc"
"dsqssa\"vzrf\"sewbp\\\\ahhlmhbeihlh"
"qtgmjck\"n\"guki\"gmdivwqxismqj"
"\"f"
"wuorvlovucngbzdszqpikyk"
"dfrdsacoukmgvhbq\"\"iwto"
"\"ey\"ch\\\\wcgioe\\\\\"ouvligmsw"
"ciqlszzgs"
"\\\\tzyrkaoi\"sopjaq"
"lmtnv"
"ar\"fqoroigiertjjlm\"ymgi\\\\kkjewsxd"
"wehcimlvudpxtamdn\"rwy"
"hr\"zvrwthr\"vruzqfrldn\"b"
"sggekodkiwvym\"mhsco"
"ltlkfbrrdvk\\\\"
"uut\"sfjnz\"\\\\ef"
"hxilg\\\\"
"zsredsiwlzrpedibn"
"vtfi"
"\\\\h"
"qekfrc\xf6wduodbwrguqcng\\\\n"
"\"lljlfdrxftwidn\\\\pkv\xd9ij"
"mrvgqynpehkliuijlpp"
"gikjph"
"yoxcdrdt\"wbaurnyhoyxoihu"
"onmomwuxuammbzxe"
"rnrr\\\\twviz\x61gqaljr\x0dmtw"
"r\"vupaoi"
"l"
"sei"
"jwxtdtbkd\\\\kxd"
"\x22v\\\\"
"ahd"
"j\"bjqxs"
"\\\\i\x24gglxub\\\\nzsajokt"
"lviwpu\"uxdlh\\\\zuy\"xqy\"ytdzlx\"r"
"kptfmys"
"fwxzikfhczkjwyjszqdbkepaeellc"
"nlqpsvbrbd\\\\ns"
"qryuwkjiodw\"\"vaqyq\"dmyifm"
"tw\x15kdmaudjl\\\\zorhp\"alwh"
"aatrvczesykekkjfyb\"kb"
"usqcutbqbxxhucwxo\xc1ltb\"j\"bghjcvws"
"ilhsrnzxkz"
"bianqfdfdhvw"
"hqibqs\x7ax\"qoxqoaqtcsz"
"htxtoojbbauztwxuiq\\\\ngyfy\\\\obzc"
"rxn\\\\moxlj"
"mtus\x84erh\"dbe"
"asx\x50huvsitcxadt"
"\"bugggtnrc\"\"kl\"hmpu\x83hqrvhpo"
"ewisbp\"\"vuzf\\\\w\x5fvalszdhl"
"scusplpwxfnxu\x57\"zynpn\x99xerc\\\\ri"
"m\\\\kinmkke\x0cl"
"xhuzit\x7fd"
"kfbo\x04\x50ruqirn"
"t\"\"xpbdscmdoug"
"punvpsgnbgyxe\"sptmpz"
"bxukkazijr"
"nxyrcdaoo\"rjkk\"wntehcvcip\"vrd"
"rdpvqskmihqaw"
"p\\\\gwdhtqnpgthod"
"nwnuf\"\"yebycearom\"nqym\"\xd4sii\xccle"
"alda\"ptspo\"wkkv\"zoi\"hkb\"vnntyd"
"ixpgpfzbqv"
"znui\"\\\\fzn\x03qozabh\"rva\"pv\x67"
"e\"zswmwuk"
"hcccygwfa"
"ngmace\\\\rtyllolr\"\x68bw"
"\\\\c\"jyufbry\"ryo\"xpo\x26ecninfeckh\\\\s"
"hdnpngtuc\"dzbvvosn\x31fwtpzbrt"
"hesbpd\xd4"
"dsdbstuzrdfmrnyntufs\"dmv"
"d\xeeibcwhcvkt"
"fvzwrsfjdqdmy\"\"v"
"ns\"dqafz\\\\lkyoflnazv\"mn\x37\"o\"yj\"e"
"dypilgbwzccayxa\"bnmuernx"
"q\xa9ztqrhreb\"\"kxfeyodqb"
"iz\xa5qjxqulaawuwz\"rqmpcj\\\\yel"
"z\"\\\\pq\"\"y\x67zpjtn"
"ifxqvivp\"kiiftdoe"
"jxzebj\"\x35\"qr\"ecglcutuoyywqumcs\"kk"
"q"
"yob\x85qmpuwexptczbkrl"
"cjiavv\"uudpozvibyycnmxhxpxmpjoz"
"xro\\\\uiqyrcid"
"nod\\\\k"
"d\"neiec"
"tqyrqvwyvmz\\\\pzgzzcqsqsrgbqbtapoz"
"r\"xvocpeuxfxslgueb\x05kzyyie\"aoec"
"\"du\\\\uirlhcbgv\\\\cjqhfreqnvn"
"zp\x04\x15\"pbjwhrjtmiba"
"\\\\cv\""
"k\"rwnb\\\\hiu\"rqd\"rc\\\\nyakrhly"
"klrmafjzandiddodgz"
"xipzhqzhvlpykzcuppx"
"zdvrvn\xd0mtfvpylbn\\\\\\\\sxcznrzugwznl"
"ody\\\\pvm\"kpjiudzhxazirgxzvumeat\"o"
"kllvhdp\"prjikzrrc\"adgpegc\\\\\"gk"
"sqtpug\xbcaauxaamw"
"wegxxrrxdvpivrqievfeokmnojsk"
"\\\\bo"
"gijhz"
"ylowluvabwrigssdgtxdwsiorxev\xdd"
"\""
"ghnsrnsqtxpygikahkrl"
"\"rcfqkbjf\"sgxg\"vnd\\\\rotn"
"ap\"smgsuexjrbuqs\"mpbstogj\"x"
"koaunz\\\\sgt\"opv"
"yialiuzwix"
"yp\"ndxgwzml\"bt"
"lpcjxmggfsy\\\\szbxccarjkqzasqkb\xcfd\x0c"
"x"
"mgakc"
"vjieunoh\x73fjwx"
"erbvv\"qulsd"
"mimycrbfhqkarmz"
"tihfbgcszuej\"c\xfbvoqskkhbgpaddioo"
"mziavkwrmekriqghw"
"izk\\\\tnjd\\\\ed\\\\emokvjoc"
"c\"nhbqzndro\\\\g"
"usfngdo"
"aypljdftvptt"
"ym\"afvq\xbcc"
"zabi\"wjpvugwhl"
"ebvptcjqjhc\"n\"p\"dxrphegr\\\\"
"mzlqqxokhye\xd9\\\\rffhnzs"
"hnipqknwpsjakanuewe"
"rqgbfcjdrmiz\"h"
"kzzp\\\\z\\\\txmkwaouxictybwx"
"yzmspjkqrteiydswlvb"
"gjpxklgpzv\"txri\\\\hotpuiukzzzd"
"p\"rxergtbsxmjmkeeqwvoagnki\""
"santipvuiq"
"\"ihjqlhtwbuy\"hdkiv\"mtiqacnf\\\\"
"oliaggtqyyx"
"fwwnpmbb"
"yrtdrieazfxyyneo"
"nywbv\\\\"
"twc\\\\ehfqxhgomgrgwpxyzmnkioj"
"qludrkkvljljd\\\\xvdeum\x4e"^;
}