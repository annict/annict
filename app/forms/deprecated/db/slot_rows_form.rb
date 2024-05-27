# typed: false
# frozen_string_literal: true

module Deprecated::Db
  class SlotRowsForm
    include ActiveModel::Model
    T.unsafe(self).include Virtus.model
    include ResourceRows

    row_model Slot

    attr_accessor :work

    validate :valid_time
    validate :valid_resource

    def set_default_rows_by_programs(program_ids, time_zone: "Asia/Tokyo")
      program_ids.split(",").each do |program_id|
        set_default_rows_by_program(program_id, time_zone)
      end
    end

    def reset_number!
      Program.where(id: attrs_list.pluck(:program_id).uniq).each do |pd|
        pd.slots.without_deleted.order(:started_at, :number).each_with_index do |p, i|
          p.update_column(:number, i + pd.minimum_episode_generatable_number)
        end
      end
    end

    private

    def set_default_rows_by_program(program_id, time_zone)
      program = @work.programs.without_deleted.find_by(id: program_id)
      return unless program

      last_slot = program.slots.without_deleted.order(started_at: :desc).first
      base_started_at = if last_slot
        last_slot.started_at + 7.days
      else
        program.started_at
      end

      rows = []
      14.times do |i|
        rows << [
          program.id,
          (base_started_at + (i * 7).days).in_time_zone(time_zone).strftime("%Y-%m-%d %H:%M")
        ]
      end

      self.rows = (self.rows || "") + rows.map { |r|
        r.join(",")
      }.join("\n") + "\n"
    end

    def attrs_list
      @attrs_list ||= fetched_rows.map { |row_data|
        {
          work_id: @work.id,
          channel_id: row_data[:channel][:id],
          episode_id: row_data[:episode][:id],
          program_id: row_data[:program][:id],
          started_at: row_data[:started_at][:value],
          rebroadcast: row_data[:rebroadcast][:value],
          time_zone: @user.time_zone
        }
      }
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        program = Program.without_deleted.find_by(id: row_columns[0])
        episode = @work.episodes.without_deleted.find_by(id: row_columns[2])

        {
          channel: {id: program&.channel&.id, value: row_columns[0]},
          started_at: {value: row_columns[1]},
          episode: {id: episode&.id},
          rebroadcast: {value: program&.rebroadcast? == true},
          program: {id: program&.id, value: row_columns[0]}
        }
      end
    end

    def valid_time
      fetched_rows.each do |row_data|
        started_at = row_data[:started_at][:value]

        begin
          Time.parse(started_at)
        rescue ArgumentError
          i18n_path = "activemodel.errors.forms.db/slot_rows_form.invalid_start_time"
          errors.add(:rows, I18n.t(i18n_path))
        end
      end
    end

    def valid_resource
      fetched_rows.each do |row_data|
        row_data.slice(:channel, :program).each do |_, data|
          next if data[:id]

          i18n_path = "activemodel.errors.forms.db/slot_rows_form.invalid"
          errors.add(:rows, I18n.t(i18n_path, value: data[:value]))
        end
      end
    end
  end
end
