# typed: false
# frozen_string_literal: true

module ActiveParameter
  extend ActiveSupport::Concern

  BASIC_ATTRS = %i[controller action access_token].freeze

  included do
    include ActiveModel::Model

    # rubocop:disable Lint/AmbiguousOperator
    attr_accessor *BASIC_ATTRS
    # rubocop:enable Lint/AmbiguousOperator

    define_model_callbacks :initialize

    def initialize(params = {})
      @params = params
      run_callbacks :initialize do
        super @params.permit(*BASIC_ATTRS)
      end
    end

    def self.param(attr, default: nil)
      method_name = "set_default_#{attr}"

      attr_accessor attr.to_sym
      after_initialize method_name.to_sym

      define_method(method_name) do
        return send(attr) if send(attr).present?

        default_value = default.is_a?(Proc) ? default.call : default
        send("#{attr}=", (@params[attr].presence || default_value))
      end

      private method_name
    end

    def fields_contain?(field)
      # `ActiveParameter` をincludeしたクラス内で `param :fields` を
      # 宣言していないとき、`fields` メソッドが未定義状態になるため、
      # `try(:fields)` としている
      try(:fields).blank? || fields.split(",").include?(field)
    end
  end
end
