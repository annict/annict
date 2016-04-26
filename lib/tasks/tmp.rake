# frozen_string_literal: true

namespace :tmp do
  task update_format: :environment do
    NumberFormat.find(1).update(data: [], format: "第%d話")
    NumberFormat.find(4).update(data: [], format: "Episode %d")
    NumberFormat.find(5).update(data: [], format: "episode %01d")
    NumberFormat.find(6).update(data: [], format: "episode.%d")
    NumberFormat.find(7).update(data: [], format: "PHASE %d")
    NumberFormat.find(8).update(data: [], format: "%d時限目")
    NumberFormat.create(name: "第一章", data: ["第一章","第二章","第三章","第四章","第五章","第六章","第七章","第八章","第九章","第十章","第十一章","第十二章","第十三章","第十四章","第十五章","第十六章","第十七章","第十八章","第十九章","第二十章","第二十一章","第二十二章","第二十三章","第二十四章","第二十五章","第二十六章","第二十七章","第二十八章","第二十九章","第三十章"])
    NumberFormat.create(name: "File.01", format: "File.%01d")
    NumberFormat.create(name: "File 1", format: "File %d")
    NumberFormat.create(name: "第一回", data: ["第一回","第二回","第三回","第四回","第五回","第六回","第七回","第八回","第九回","第十回","第十一回","第十二回","第十三回","第十四回","第十五回","第十六回","第十七回","第十八回","第十九回","第二十回","第二十一回","第二十二回","第二十三回","第二十四回","第二十五回","第二十六回","第二十七回","第二十八回","第二十九回","第三十回"])
    NumberFormat.create(name: "第1回", format: "第%d回")
    NumberFormat.create(name: "第01回", format: "第%01d回")
    NumberFormat.create(name: "Mission 01", format: "Mission %01d")
    NumberFormat.create(name: "LV.01", format: "LV.%01d")
    NumberFormat.create(name: "#01", format: "#%01d")
    NumberFormat.create(name: "第1幕", format: "第%d幕")
  end
end
