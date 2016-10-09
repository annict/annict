# frozen_string_literal: true

require "csv"

module ResourceRows
  extend ActiveSupport::Concern

  included do
    def parsed_rows
      # Replace double quote to avoid `CSV::MalformedCSVError`
      rows = self.rows.gsub(/"/, "__double_quote__")

      CSV.parse(rows).map do |row_columns|
        row_columns.map { |column| column&.gsub("__double_quote__", '"')&.strip }
      end
    end
  end
end
