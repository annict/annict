# frozen_string_literal: true

class Deprecated::EpisodeGeneratorService
  def self.execute!(now: Time.current)
    new.execute!(now: now)
  end

  def execute!(now:)
    episodeless_slots(now).order(:started_at).each do |slot|
      work = slot.work

      next if work.manual_episodes_count && work.manual_episodes_count < slot.number

      raw_number = target_raw_number(slot)
      episode = work.episodes.only_kept.find_by(raw_number: raw_number)

      if episode
        slot.update_column(:episode_id, episode.id)
        next
      end

      create_new_episode!(slot, raw_number)
    rescue => e
      AdminMailer.error_in_episode_generator_notification(slot.id, e.message).deliver_later
    end
  end

  private

  def episodeless_slots(now)
    Slot
      .only_kept
      .where(episode_id: nil, rebroadcast: false, irregular: false)
      .where.not(program_id: nil)
      .where.not(number: nil)
      .before(now + 7.days, field: :started_at)
      .includes(:work)
  end

  def target_raw_number(slot)
    irregular_slots_count = Slot
      .only_kept
      .where(program_id: slot.program_id, irregular: true)
      .where("number >= ?", slot.program.minimum_episode_generatable_number)
      .count
    raw_number = slot.number - irregular_slots_count
    raw_number + slot.work.start_episode_raw_number - 1
  end

  def create_new_episode!(slot, raw_number)
    work = slot.work
    new_episode = nil

    ActiveRecord::Base.transaction do
      new_episode = work.episodes.create!(
        raw_number: raw_number,
        number: work.formatted_number(raw_number).presence || "第#{raw_number.to_i}話",
        sort_number: raw_number.to_i * 100
      )

      slot.update_column(:episode_id, new_episode.id)
    end

    AdminMailer.episode_created_notification(new_episode&.id).deliver_later
  end
end
