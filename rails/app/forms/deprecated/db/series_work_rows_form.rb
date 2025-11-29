# typed: false
# frozen_string_literal: true

module Deprecated::Db
  class SeriesWorkRowsForm
    include ActiveModel::Model
    T.unsafe(self).include Virtus.model
    include ResourceRows

    attr_accessor :series

    row_model SeriesWork

    validate :valid_resource

    private

    def attrs_list
      @attrs_list ||= fetched_rows.map { |row_data|
        {
          series_id: @series.id,
          work_id: row_data[:work][:id],
          summary: row_data[:summary][:value]
        }
      }
    end

    def fetched_rows
      parsed_rows.map do |row_columns|
        work = Work.only_kept.where(id: row_columns[0])
          .or(Work.only_kept.where(title: row_columns[0])).first

        {
          work: {id: work&.id, value: row_columns[0]},
          summary: {value: row_columns[1].presence || ""}
        }
      end
    end

    def valid_resource
      fetched_rows.each do |row_data|
        row_data.slice(:work).each do |_, data|
          next if data[:id].present?
          i18n_path = "activemodel.errors.forms.db/series_work_rows_form.invalid"
          errors.add(:rows, I18n.t(i18n_path, value: data[:value]))
        end
      end
    end
  end
end
