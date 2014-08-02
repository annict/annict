require 'open-uri'

namespace :syobocal do
  task save: :environment do
    [
      :save_channel_groups,
      :save_channels,
      :save_sc_tid_on_works,
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

  task save_sc_tid_on_works: :environment do
    doc = Nokogiri::XML(open('http://cal.syoboi.jp/db.php?Command=TitleLookup&TID=*&Fields=TID,Title,Cat'))
    doc.css('TitleItem').each do |item|
      tid      = item.xpath('TID').text.to_i
      title    = item.xpath('Title').text
      # タイトルのカテゴリ値 https://sites.google.com/site/syobocal/spec/title-cat
      category = item.xpath('Cat').text.to_i

      # アニメ、アニメ(終了/再放送)、OVA、映画だったら
      if [1, 10, 7, 8].include?(category)
        work = Work.find_by(title: title)

        if work.present? && work.sc_tid.blank?
          work.update_column(:sc_tid, tid)

          puts "workを更新しました。#{work.title}: #{work.sc_tid}"
        end
      end
    end
  end

  task save_programs: :environment do
    tids        = get_work_sc_tids
    fields      = 'TID,StTime,Count,Deleted,ChID,STSubTitle,LastUpdate'
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
      tid   = item.xpath('TID').text.to_i
      title = item.xpath('STSubTitle').text
      count = item.xpath('Count').text.to_i
      work  = Work.find_by(sc_tid: tid)

      if work.present? && (count.present? && count >= 1)
        episode = work.episodes.find_by(title: title).presence || work.episodes.find_by(sc_count: count)
        title   = title.presence || '-'

        if episode.present?
          if episode.title != title || episode.sc_count.blank?
            episode.update_attributes(title: title, sc_count: count)

            puts "episodeを更新しました。id: #{episode.id}, number: #{episode.number}, episode_title: #{episode.title}"
          end
        else
          episode = work.episodes.create do |e|
            e.number      = "##{count}"
            e.sort_number = work.episodes.count + 1
            e.sc_count    = count
            e.title       = title
          end

          puts "episodeを作成しました。id: #{episode.id}, number: #{episode.number}, episode_title: #{episode.title}"
          SyobocalMailer.delay.episode_created_notification(episode.id)
        end

        chid    = item.xpath('ChID').text.to_i
        channel = Channel.find_by(sc_chid: chid)

        #「ニコニコチャンネル (165 == chid)」に対しての番組情報は別の手段で登録する
        if 165 != chid && channel.present?
          program     = channel.programs.find_by(episode_id: episode.id)
          st_time     = DateTime.parse(item.xpath('StTime').text).in_time_zone('Asia/Tokyo') - 9.hours
          last_update = DateTime.parse(item.xpath('LastUpdate').text).in_time_zone('Asia/Tokyo') - 9.hours

          if program.present?
            if last_update > program.sc_last_update
              deleted = item.xpath('Deleted').text == '1' ? true : false

              if deleted
                program.destroy
                puts 'programを削除しました。'
              else
                program.update_attributes(sc_last_update: last_update, started_at: st_time)
                puts 'programを更新しました。'
              end
            end
          else
            program = channel.programs.create(episode_id: episode.id, work_id: episode.work.id, sc_last_update: last_update, started_at: st_time)
            puts "programを作成しました。channel: #{program.channel.name}, work: #{program.work.title}, episode: #{program.episode.number} #{program.episode.title}"
          end
        end
      end
    end
  end

  def get_work_sc_tids
    work_sc_tids = Work.where(fetch_syobocal: true).pluck(:sc_tid).select { |sid| sid.present? }

    work_sc_tids.to_s.gsub(/\[|\]| /, '')
  end
end