module Syobocal
  class ProgItem
    def initialize(item)
      @item = item
      @pid = item_text('PID').to_i
      @tid = item_text('TID').to_i
      @title = item_text('STSubTitle')
      @sub_title = item_text('SubTitle')
      @count = item_text('Count').to_i
      @chid = item_text('ChID').to_i
      @st_time = item_datetime('StTime')
      @last_update = item_datetime('LastUpdate')
      @deleted = item_text('Deleted').to_i
      @flag = item_text('Flag').to_i
      @prog_comment = item_text('ProgComment')
    end

    def work
      @work ||= Work.find_by(sc_tid: @tid)
    end

    def episode
      episode_by_sc_count = work.episodes.find_by(sc_count: @count)
      episode_by_title = @title.blank? ? nil : work.episodes.find_by(title: @title)
      @episode ||= (episode_by_sc_count.presence || episode_by_title)
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

    def rebroadcast_program?
      @flag.in?([8, 9, 10, 11, 12, 13])
    end

    def save_episode
      title = @title.presence || nil

      episode = self.episode.present? ? update_episode(title) : create_episode(title)
      episode
    end

    def save_program(episode)
      # 「ニコニコチャンネル」に対しての放送予定は別の手段で登録する
      if !niconico_ch? && channel.present?
        program = Program.find_by(sc_pid: @pid) ||
          channel.programs.find_by(episode_id: episode.id)

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
      end
    end

    private

    def item_text(name)
      @item.xpath(name).text
    end

    def item_datetime(name)
      DateTime.parse(item_text(name)).in_time_zone('Asia/Tokyo') - 9.hours
    end

    def create_episode(title)
      prev_episode = work.episodes.order(sort_number: :desc).first
      episode = work.episodes.new do |e|
        e.number = formatted_number(@count).presence || "##{@count}"
        e.raw_number = @count
        e.sort_number = (work.episodes.count + 1) * 10
        e.sc_count = @count
        e.title = title
        e.fetch_syobocal = true
      end
      episode.prev_episode_id = prev_episode.id if prev_episode.present?

      episode.save!

      SyobocalMailer.delay.episode_created_notification(episode.id)

      episode
    end

    def update_episode(title)
      if episode.fetch_syobocal? && (episode.sc_count.blank? || episode.title != title)
        episode.update_attributes(title: title, sc_count: @count)
      end

      episode
    end

    def formatted_number(count)
      return nil if work.number_format.blank?
      work.number_format.data[count - 1]
    end

    def create_program(episode)
      program = channel.programs.create do |p|
        p.episode_id = episode.id
        p.work_id = episode.work.id
        p.sc_pid = @pid
        p.rebroadcast = rebroadcast_program?
        p.sc_last_update = @last_update
        p.started_at = @st_time
      end
    end

    def update_program(program)
      if @last_update > program.sc_last_update
        if deleted?
          program.destroy
        else
          program.update_attributes(sc_last_update: @last_update, started_at: @st_time)
        end
      end
    end
  end
end
