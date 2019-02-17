# frozen_string_literal: true

module DB
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

    private

    def attrs_list
      @attrs_list ||= fetched_rows.map do |row_data|
        {
          work_id: @work.id,
          channel_id: row_data[:channel][:id],
          episode_id: row_data[:episode][:id],
          started_at: row_data[:started_at][:value],
          rebroadcast: row_data[:rebroadcast][:value],
          time_zone: @user.time_zone
        }
      end
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        channel = Channel.published.where(id: row_columns[0]).
          or(Channel.published.where(name: row_columns[0])).first
        episode = @work.episodes.published.where(id: row_columns[1]).
          or(@work.episodes.published.where(number: row_columns[1])).
          or(@work.episodes.published.where(title: row_columns[1])).first

        {
          channel: { id: channel&.id, value: row_columns[0] },
          episode: { id: episode&.id, value: row_columns[1] },
          started_at: { value: row_columns[2] },
          rebroadcast: { value: row_columns[3] == "1" }
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
        row_data.slice(:channel).each do |_, data|
          next if data[:id]

          i18n_path = "activemodel.errors.forms.db/program_rows_form.invalid"
          errors.add(:rows, I18n.t(i18n_path, value: data[:value]))
        end
      end
    end
  end
end
