# frozen_string_literal: true

class EpisodeGeneratorService
  def self.execute!(now: Time.current)
    new.execute!(now: now)
  end

  def execute!(now:)
    episodeless_programs = Program.
      published.
      where(episode_id: nil, rebroadcast: false).
      where.not(number: nil).
      before(now + 7.days, field: :started_at).
      includes(:work)

    episodeless_programs.order(:started_at).each do |p|
      work = p.work

      next if work.manual_episodes_count && work.manual_episodes_count < p.number

      episode = work.episodes.published.order(:sort_number)[p.number - 1]
      episode_not_exists = episode.nil?
      ActiveRecord::Base.transaction do
        new_number = p.number - work.irregular_episodes_count
        episode ||= work.episodes.create!(
          raw_number: new_number,
          number: "第#{new_number}話",
          sort_number: p.number * 100
        )

        p.update_column(:episode_id, episode.id)
      end

      AdminMailer.episode_created_notification(episode.id).deliver_later if episode_not_exists
    end
  end
end
