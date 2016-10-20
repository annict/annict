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

    private

    def attrs_list
      @attrs_list ||= fetched_rows.map do |row_data|
        {
          work_id: @work.id,
          channel_id: row_data[:channel][:id],
          episode_id: row_data[:episode][:id],
          started_at: row_data[:started_at][:value],
          rebroadcast: row_data[:rebroadcast][:value].presence || false,
          time_zone: @user.time_zone
        }
      end
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        channel = Channel.where(id: row_columns[0]).
          or(Channel.where(name: row_columns[0])).first
        episode = @work.episodes.where(number: row_columns[1]).
          or(@work.episodes.where(title: row_columns[1])).first

        {
          channel: { id: channel&.id, value: row_columns[0] },
          episode: { id: episode&.id, value: row_columns[1] },
          started_at: { value: row_columns[2] },
          rebroadcast: { value: row_columns[3] }
        }
      end
    end

    def valid_time
      fetched_rows.each do |row_data|
        started_at = row_data[:started_at][:value]

        begin
          Time.parse(started_at)
        rescue
          i18n_path = "activemodel.errors.forms.db/program_rows_form.invalid_start_time"
          errors.add(:rows, I18n.t(i18n_path))
        end
      end
    end

    def valid_resource
      fetched_rows.each do |row_data|
        row_data.slice(:channel, :episode).each do |_, data|
          next if data[:id].present?
          i18n_path = "activemodel.errors.forms.db/program_rows_form.invalid"
          errors.add(:rows, I18n.t(i18n_path, value: data[:value]))
        end
      end
    end
  end
end
