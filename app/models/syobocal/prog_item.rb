module Syobocal
  class ProgItem
    def initialize(item)
      @pid = item.xpath('PID').text.to_i
      @tid = item.xpath('TID').text.to_i
      @title = item.xpath('STSubTitle').text
      @sub_title = item.xpath('SubTitle').text
      @count = item.xpath('Count').text.to_i
      @chid = item.xpath('ChID').text.to_i
      @st_time = DateTime.parse(item.xpath('StTime').text).in_time_zone('Asia/Tokyo') - 9.hours
      @last_update = DateTime.parse(item.xpath('LastUpdate').text).in_time_zone('Asia/Tokyo') - 9.hours
      @deleted = item.xpath('Deleted').text.to_i
      @flag = item.xpath('Flag').text.to_i
      @prog_comment = item.xpath('ProgComment').text
    end

    def work
      @work ||= Work.find_by(sc_tid: @tid)
    end

    def episode
      @episode ||= (work.episodes.find_by(title: @title).presence || work.episodes.find_by(sc_count: @count))
    end

    def channel
      @channel ||= Channel.find_by(sc_chid: @chid)
    end

    def normal_program?
      work.present? && episode?
    end

    def special_program?
      /^#[0-9]+/ === @sub_title
    end

    def episode?
      @count.present? && @count >= 1
    end

    def niconico_ch?
      @chid == 165
    end

    def deleted?
      @deleted == 1
    end

    def new_program?
      @flag == 2
    end

    def end_program?
      @flag == 4
    end

    def save_episode
      title = @title.presence || '-'

      episode = self.episode.present? ? update_episode(title) : create_episode(title)
      episode
    end

    def save_program(episode)
      #「ニコニコチャンネル」に対しての番組情報は別の手段で登録する
      if !niconico_ch? && channel.present?
        program = channel.programs.find_by(episode_id: episode.id)

        if program.present?
          update_program(program)
        else
          create_program(episode)
        end
      end
    end

    def save_alert
      if special_program? && Syobocal::Alert.new_special_program?(@pid)
        alert = Syobocal::Alert.create do |a|
          a.work_id = work.id
          a.kind = :special_program
          a.sc_prog_item_id = @pid
          a.sc_sub_title = @sub_title
          a.sc_prog_comment = @prog_comment
        end

        SyobocalMailer.delay.special_program_notification(alert.id)

        message = "alertを作成しました。id: #{alert.id}, kind: #{alert.kind_text}, sc_sub_title: #{alert.sc_sub_title}"
        puts message
        Rails.logger.info(message)
      end
    end


    private

    def create_episode(title)
      episode = work.episodes.create do |e|
        e.number = "##{@count}"
        e.sort_number = (work.episodes.count + 1) * 10
        e.sc_count = @count
        e.title = title
      end

      SyobocalMailer.delay.episode_created_notification(episode.id)

      message = "episodeを作成しました。id: #{episode.id}, number: #{episode.number}, episode_title: #{episode.title}"
      puts message
      Rails.logger.info(message)

      episode
    end

    def update_episode(title)
      if episode.sc_count.blank? || episode.title != title
        episode.update_attributes(title: title, sc_count: @count)

        message = "episodeを更新しました。id: #{episode.id}, number: #{episode.number}, episode_title: #{episode.title}"
        Rails.logger.info(message)
        puts message
      end

      episode
    end

    def create_program(episode)
      program = channel.programs.create do |p|
        p.episode_id = episode.id
        p.work_id = episode.work.id
        p.sc_last_update = @last_update
        p.started_at = @st_time
      end

      message = "programを作成しました。channel: #{program.channel.name}, work: #{program.work.title}, episode: #{program.episode.number} #{program.episode.title}"
      Rails.logger.info(message)
      puts message
    end

    def update_program(program)
      if @last_update > program.sc_last_update
        if deleted?
          program.destroy

          message = "programを削除しました。channel: #{program.channel.name}, work: #{program.work.title}, episode: #{program.episode.number} #{program.episode.title}"
        else
          program.update_attributes(sc_last_update: @last_update, started_at: @st_time)

          message = "programを更新しました。channel: #{program.channel.name}, work: #{program.work.title}, episode: #{program.episode.number} #{program.episode.title}"
        end

        Rails.logger.info(message)
        puts message
      end
    end
  end
end
