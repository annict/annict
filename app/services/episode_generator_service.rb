# frozen_string_literal: true

class EpisodeGeneratorService
  def self.execute!(now: Time.current)
    new.execute!(now: now)
  end

  def execute!(now:)
    episodeless_programs = Program.
      published.
      where(episode_id: nil, rebroadcast: false, irregular: false).
      where.not(program_detail_id: nil).
      where.not(number: nil).
      before(now + 7.days, field: :started_at).
      includes(:work)

    episodeless_programs.order(:started_at).each do |p|
      work = p.work

      next if work.manual_episodes_count && work.manual_episodes_count < p.number

      irregular_programs_count = Program.
        published.
        where(program_detail_id: p.program_detail_id, irregular: true).
        where("number >= ?", p.program_detail.minimum_episode_generatable_number).
        count
      raw_number = p.number - irregular_programs_count
      raw_number = raw_number + work.start_episode_raw_number - 1
      episode = work.episodes.published.find_by(raw_number: raw_number)
      episode_not_exists = episode.nil?

      ActiveRecord::Base.transaction do
        episode ||= work.episodes.create!(
          raw_number: raw_number,
          number: work.formatted_number(raw_number).presence || "第#{raw_number}話",
          sort_number: p.number * 100
        )

        p.update_column(:episode_id, episode.id)
      end

      AdminMailer.episode_created_notification(episode.id).deliver_later if episode_not_exists
    end
  end
end
