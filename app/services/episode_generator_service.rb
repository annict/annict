# frozen_string_literal: true

class EpisodeGeneratorService
  def self.execute!(now: Time.current)
    new.execute!(now: now)
  end

  def execute!(now:)
    episodeless_slots = Slot.
      without_deleted.
      where(episode_id: nil, rebroadcast: false, irregular: false).
      where.not(program_id: nil).
      where.not(number: nil).
      before(now + 7.days, field: :started_at).
      includes(:work)

    episodeless_slots.order(:started_at).each do |slot|
      work = slot.work

      next if work.manual_episodes_count && work.manual_episodes_count < slot.number

      irregular_slots_count = Slot.
        without_deleted.
        where(program_id: slot.program_id, irregular: true).
        where("number >= ?", slot.program.minimum_episode_generatable_number).
        count
      raw_number = slot.number - irregular_slots_count
      raw_number = raw_number + work.start_episode_raw_number - 1
      episode = work.episodes.without_deleted.find_by(raw_number: raw_number)
      episode_not_exists = episode.nil?

      ActiveRecord::Base.transaction do
        episode ||= work.episodes.create!(
          raw_number: raw_number,
          number: work.formatted_number(raw_number).presence || "第#{raw_number.to_i}話",
          sort_number: slot.number * 100
        )

        slot.update_column(:episode_id, episode.id)
      end

      AdminMailer.episode_created_notification(episode.id).deliver_later if episode_not_exists
    end
  end
end
