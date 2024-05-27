# typed: false
# frozen_string_literal: true

require "csv"

module ResourceRows
  extend ActiveSupport::Concern

  included do
    attr_accessor :user

    attribute :rows, String

    validates :rows, presence: true

    def self.row_model(model)
      define_method :new_resources do
        @new_resources ||= attrs_list.map { |attrs| model.new(attrs) }
      end
    end

    def valid?
      super && new_resources.all?(&:valid?)
    end

    def save!
      new_resources_with_user.each(&:save_and_create_activity!)
    end

    private

    def parsed_rows
      # Replace double quote to avoid `CSV::MalformedCSVError`
      rows = self.rows.gsub(/"/, "__double_quote__")

      CSV.parse(rows).reject(&:empty?).map do |row_columns|
        row_columns.map { |column| column&.gsub("__double_quote__", '"')&.strip }
      end
    end

    def new_resources_with_user
      @new_resources_with_user ||= new_resources.map { |resource|
        resource.user = @user
        resource
      }
    end
  end
end
