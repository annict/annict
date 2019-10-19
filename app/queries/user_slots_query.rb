# frozen_string_literal: true

class UserSlotsQuery
  attr_reader :user

  def initialize(user)
    @user = user
  end

  # 記録していないエピソードと紐づく番組情報を返す
  def unwatched_all
    Slot.published.where(id: slot_ids(channel_works, scope: :unwatched))
  end

  def all
    Slot.published.where(id: slot_ids(channel_works, scope: :all))
  end

  def unwatched(page, sort)
    unwatched_all.
      work_published.
      episode_published.
      where("started_at < ?", Date.tomorrow + 1.day + 5.hours).
      includes(:channel, work: [:work_image], episode: [:work]).
      order(started_at: sort_type(sort)).
      page(page)
  end

  private

  def channel_works
    works = user.works.wanna_watch_and_watching
    user.channel_works.includes(:work).where(work: works)
  end

  def slot_ids(channel_works, scope: :all)
    slot_ids = []

    channel_works.each do |cw|
      episode_ids = case scope
      when :all then cw.work.episodes.published.pluck(:id)
      when :unwatched then user.episodes.unwatched(cw.work).pluck(:id)
      end

      next if episode_ids.blank?

      # 過去の放送を取得しないようにするため、直近の放送分のみ取得する
      sql = <<-SQL
        WITH ranked_slots AS (
          SELECT id, episode_id, started_at,
            dense_rank() OVER (
              PARTITION BY episode_id ORDER BY started_at ASC
            ) AS episode_rank
          FROM slots
          WHERE
            channel_id = #{cw.channel_id} AND
            episode_id IN (#{episode_ids.join(",")})
        )
        SELECT id FROM ranked_slots WHERE
          episode_rank = (SELECT max(episode_rank) FROM ranked_slots);
      SQL

      slot_ids << Slot.published.find_by_sql(sql).map(&:id)
    end

    slot_ids.flatten
  end

  def sort_type(sort)
    return :asc if sort == "started_at_asc"
    return :desc if sort == "started_at_desc"
    :desc
  end
end
