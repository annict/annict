# frozen_string_literal: true

module Db
  class ProgramRowsForm
    include ActiveModel::Model
    include Virtus.model
    include ResourceRows

    row_model Program

    attr_accessor :work

    validate :valid_time
    validate :valid_resource

    def set_default_rows_by_program_detail!(program_detail_id, time_zone: "Asia/Tokyo")
      program_detail = @work.program_details.published.find_by(id: program_detail_id)
      return unless program_detail

      last_program = program_detail.programs.published.order(started_at: :desc).first
      base_started_at = if last_program
        last_program.started_at + 7.days
      else
        program_detail.started_at
      end

      rows = []
      14.times do |i|
        rows << [
          program_detail.id,
          (base_started_at + (i * 7).days).in_time_zone(time_zone).strftime("%Y-%m-%d %H:%M"),
        ]
      end

      self.rows = rows.map do |r|
        r.join(",")
      end.join("\n")
    end

    def reset_number!
      ProgramDetail.where(id: attrs_list.pluck(:program_detail_id).uniq).each do |pd|
        pd.programs.published.order(:started_at).each_with_index do |p, i|
          p.update_column(:number, i + pd.minimum_episode_generatable_number)
        end
      end
    end

    private

    def attrs_list
      @attrs_list ||= fetched_rows.map do |row_data|
        {
          work_id: @work.id,
          channel_id: row_data[:channel][:id],
          episode_id: row_data[:episode][:id],
          program_detail_id: row_data[:program_detail][:id],
          started_at: row_data[:started_at][:value],
          rebroadcast: row_data[:rebroadcast][:value],
          time_zone: @user.time_zone
        }
      end
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        program_detail = ProgramDetail.published.find_by(id: row_columns[0])
        episode = @work.episodes.published.find_by(id: row_columns[2])

        {
          channel: { id: program_detail&.channel&.id, value: row_columns[0] },
          started_at: { value: row_columns[1] },
          episode: { id: episode&.id },
          rebroadcast: { value: program_detail&.rebroadcast? == true },
          program_detail: { id: program_detail&.id, value: row_columns[0] }
        }
      end
    end

    def valid_time
      fetched_rows.each do |row_data|
        started_at = row_data[:started_at][:value]

        begin
          Time.parse(started_at)
        rescue ArgumentError
          i18n_path = "activemodel.errors.forms.db/program_rows_form.invalid_start_time"
          errors.add(:rows, I18n.t(i18n_path))
        end
      end
    end

    def valid_resource
      fetched_rows.each do |row_data|
        row_data.slice(:channel, :program_detail).each do |_, data|
          next if data[:id]

          i18n_path = "activemodel.errors.forms.db/program_rows_form.invalid"
          errors.add(:rows, I18n.t(i18n_path, value: data[:value]))
        end
      end
    end
  end
end
