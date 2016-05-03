# frozen_string_literal: true

module ActiveParameter
  extend ActiveSupport::Concern

  included do
    include ActiveModel::Model

    attr_accessor :controller, :action, :access_token
    define_model_callbacks :initialize

    def initialize(params = {})
      run_callbacks :initialize do
        super params
      end
    end

    def self.param(attr, default: nil)
      method_name = "set_default_#{attr}"

      attr_accessor attr.to_sym
      after_initialize method_name.to_sym

      define_method(method_name) do
        return send(attr) if send(attr).present?

        value = default.is_a?(Proc) ? default.call : default
        send("#{attr}=", value)
      end

      private method_name
    end
  end
end
