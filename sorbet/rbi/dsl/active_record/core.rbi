# typed: true

# DO NOT EDIT MANUALLY
# This is an autogenerated file for dynamic methods in `ActiveRecord::Core`.
# Please instead update this file by running `bin/tapioca dsl ActiveRecord::Core`.


module ActiveRecord::Core
  include GeneratedInstanceMethods

  mixes_in_class_methods GeneratedClassMethods

  module GeneratedClassMethods
    def belongs_to_required_by_default; end
    def belongs_to_required_by_default=(value); end
    def belongs_to_required_by_default?; end
    def default_connection_handler; end
    def default_connection_handler=(value); end
    def default_connection_handler?; end
    def default_role; end
    def default_role=(value); end
    def default_role?; end
    def default_shard; end
    def default_shard=(value); end
    def default_shard?; end
    def destroy_association_async_job; end
    def destroy_association_async_job=(value); end
    def enumerate_columns_in_select_statements; end
    def enumerate_columns_in_select_statements=(value); end
    def enumerate_columns_in_select_statements?; end
    def has_many_inversing; end
    def has_many_inversing=(value); end
    def has_many_inversing?; end
    def logger; end
    def logger=(value); end
    def logger?; end
    def shard_selector; end
    def shard_selector=(value); end
    def shard_selector?; end
    def strict_loading_by_default; end
    def strict_loading_by_default=(value); end
    def strict_loading_by_default?; end
  end

  module GeneratedInstanceMethods
    def default_connection_handler; end
    def default_connection_handler?; end
    def default_role; end
    def default_role?; end
    def default_shard; end
    def default_shard?; end
    def destroy_association_async_job; end
    def logger; end
    def logger?; end
  end
end
