require 'open-uri'

namespace :syobocal do
  task save: :environment do
    [
      :save_channel_groups,
      :save_channels,
      :save_programs
    ].each do |task_name|
      puts "============== #{task_name} =============="
      Rake::Task["syobocal:#{task_name}"].invoke
    end
  end

  task save_channel_groups: :environment do
    doc = Nokogiri::XML(open('http://cal.syoboi.jp/db.php?Command=ChGroupLookup'))
    doc.css('ChGroupItem').each do |item|
      name = item.xpath('ChGroupName').text

      if ChannelGroup.where(name: name).blank?
        channel_group = ChannelGroup.create do |cg|
          cg.name     = name
          cg.sc_chgid = item.xpath('ChGID').text
        end

        puts "channel_groupを作成しました。name: #{channel_group.name}"
      end
    end
  end

  task save_channels: :environment do
    doc = Nokogiri::XML(open('http://cal.syoboi.jp/db.php?Command=ChLookup'))
    doc.css('ChItem').each do |item|
      sc_chgid = item.xpath('ChGID').text
      channel_group = ChannelGroup.find_by(sc_chgid: sc_chgid)

      if channel_group.present?
        conditions = {
          sc_chid: item.xpath('ChID').text,
          name:    item.xpath('ChName').text
        }

        if channel_group.channels.where(conditions).blank?
          channel = channel_group.channels.create(conditions)

          puts "channelを作成しました。name: #{channel.name}"
        end
      else
        puts 'new!!'
      end
    end
  end

  task save_programs: :environment do
    tids        = get_work_sc_tids
    fields      = 'PID,TID,StTime,Count,SubTitle,ProgComment,Flag,Deleted,ChID,STSubTitle,LastUpdate'
    started_on  = (Date.today - 5.days).strftime('%Y%m%d')
    ended_on    = (Date.today + 5.days).strftime('%Y%m%d')
    range       = "#{started_on}_000000-#{ended_on}_235959"
    request_url = "http://cal.syoboi.jp/db.php" +
                  "?Command=ProgLookup" +
                  "&TID=#{tids}" +
                  "&JOIN=SubTitles" +
                  "&Fields=#{fields}" +
                  "&Range=#{range}"
    doc = Nokogiri::XML(open(request_url))

    doc.css('ProgItem').each do |item|
      prog_item = Syobocal::ProgItem.new(item)

      if prog_item.normal_program?
        episode = prog_item.save_episode
        prog_item.save_program(episode)
      else
        prog_item.save_alert
      end
    end
  end

  def get_work_sc_tids
    work_sc_tids = Work.where(fetch_syobocal: true).pluck(:sc_tid).select { |sid| sid.present? }

    work_sc_tids.to_s.gsub(/\[|\]| /, '')
  end
end
