# frozen_string_literal: true

module Annict
  module Amazon
    class Client
      def initialize(country: "jp")
        @country = country
        @ecs ||= ::Amazon::Ecs.configure do |options|
          options[:AWS_access_key_id] = ENV.fetch("AWS_PAA_ACCESS_KEY_ID")
          options[:AWS_secret_key] = ENV.fetch("AWS_PAA_SECRET_KEY")
          options[:associate_tag] = ENV.fetch("AMAZON_ASSOCIATE_TAG")
        end
      end

      def items
        @items ||= Annict::Amazon::Items.new(country: @country)
      end
    end
  end
end
