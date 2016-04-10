# frozen_string_literal: true

namespace :tmp do
  task update_title_kana: :environment do
    tids = Work.where.not(sc_tid: nil).pluck(:sc_tid)
    url = "http://cal.syoboi.jp/db.php?Command=TitleLookup&TID=#{tids.join(',')}&Fields=TitleYomi"
    doc = Nokogiri::XML(open(url))
    doc.css("TitleItem").each do |item|
      sc_tid = item.attribute("id").value
      next if sc_tid.blank?
      title_kana = item.css("TitleYomi").text
      next if title_kana.blank?
      Work.find_by(sc_tid: sc_tid).update_column(:title_kana, title_kana)
      puts "updated. sc_tid: #{sc_tid}, title_kana: #{title_kana}"
    end
  end
end
