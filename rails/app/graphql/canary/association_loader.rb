# typed: false
# frozen_string_literal: true

module Canary
  # Based on https://github.com/Shopify/graphql-batch/blob/de001c84f20d041dc3a5964ba886ac79545d993e/examples/association_loader.rb
  class AssociationLoader < GraphQL::Batch::Loader
    def self.validate(model, association_names)
      new(model, association_names)
      nil
    end

    def initialize(model, association_names)
      @model = model
      @association_names = association_names
      validate
    end

    def load(record)
      raise TypeError, "#{model} loader can't load association for #{record.class}" unless record.is_a?(model)
      return Promise.resolve(read_associations(record)) if all_associations_loaded?(record)
      super
    end

    # We want to load the associations on all records, even if they have the same id
    def cache_key(record)
      record.object_id
    end

    def perform(records)
      preload_associations(records)
      records.each { |record| fulfill(record, read_associations(record)) }
    end

    private

    attr_reader :association_names, :model

    def validate
      association_names.each do |association_name|
        unless model.reflect_on_association(association_name)
          raise ArgumentError, "No association #{association_name} on #{model}"
        end
      end
    end

    def preload_associations(records)
      association_names.each do |association_name|
        ::ActiveRecord::Associations::Preloader.new.preload(records, association_name)
      end
    end

    def read_associations(record)
      association_names.flat_map do |association_name|
        record.public_send(association_name)
      end
    end

    def all_associations_loaded?(record)
      association_names.all? do |association_name|
        record.association(association_name).loaded?
      end
    end
  end
end
