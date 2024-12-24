# typed: true

# DO NOT EDIT MANUALLY
# This is an autogenerated file for types exported from the `globalid` gem.
# Please instead update this file by running `bin/tapioca gem globalid`.


# source://globalid//lib/global_id/global_id.rb#7
class GlobalID
  extend ::ActiveSupport::Autoload

  # @return [GlobalID] a new instance of GlobalID
  #
  # source://globalid//lib/global_id/global_id.rb#44
  def initialize(gid, options = T.unsafe(nil)); end

  # source://globalid//lib/global_id/global_id.rb#63
  def ==(other); end

  # source://globalid//lib/global_id/global_id.rb#42
  def app(*_arg0, **_arg1, &_arg2); end

  # source://globalid//lib/global_id/global_id.rb#76
  def as_json(*_arg0); end

  # source://globalid//lib/global_id/global_id.rb#42
  def deconstruct_keys(*_arg0, **_arg1, &_arg2); end

  # source://globalid//lib/global_id/global_id.rb#63
  def eql?(other); end

  # source://globalid//lib/global_id/global_id.rb#48
  def find(options = T.unsafe(nil)); end

  # source://globalid//lib/global_id/global_id.rb#68
  def hash; end

  # source://globalid//lib/global_id/global_id.rb#52
  def model_class; end

  # source://globalid//lib/global_id/global_id.rb#42
  def model_id(*_arg0, **_arg1, &_arg2); end

  # source://globalid//lib/global_id/global_id.rb#42
  def model_name(*_arg0, **_arg1, &_arg2); end

  # source://globalid//lib/global_id/global_id.rb#42
  def params(*_arg0, **_arg1, &_arg2); end

  # source://globalid//lib/global_id/global_id.rb#72
  def to_param; end

  # source://globalid//lib/global_id/global_id.rb#42
  def to_s(*_arg0, **_arg1, &_arg2); end

  # Returns the value of attribute uri.
  #
  # source://globalid//lib/global_id/global_id.rb#41
  def uri; end

  class << self
    # Returns the value of attribute app.
    #
    # source://globalid//lib/global_id/global_id.rb#9
    def app; end

    # source://globalid//lib/global_id/global_id.rb#31
    def app=(app); end

    # source://globalid//lib/global_id/global_id.rb#11
    def create(model, options = T.unsafe(nil)); end

    # source://globalid//lib/global_id.rb#20
    def deprecator; end

    # source://globalid//lib/global_id.rb#15
    def eager_load!; end

    # source://globalid//lib/global_id/global_id.rb#21
    def find(gid, options = T.unsafe(nil)); end

    # source://globalid//lib/global_id/global_id.rb#25
    def parse(gid, options = T.unsafe(nil)); end

    private

    # source://globalid//lib/global_id/global_id.rb#36
    def parse_encoded_gid(gid, options); end
  end
end

# Mix `GlobalID::Identification` into any model with a `#find(id)` class
# method. Support is automatically included in Active Record.
#
#   class Person
#     include ActiveModel::Model
#     include GlobalID::Identification
#
#     attr_accessor :id
#
#     def self.find(id)
#       new id: id
#     end
#
#     def ==(other)
#       id == other.try(:id)
#     end
#   end
#
#   person_gid = Person.find(1).to_global_id
#   # => #<GlobalID ...
#   person_gid.uri
#   # => #<URI ...
#   person_gid.to_s
#   # => "gid://app/Person/1"
#   GlobalID::Locator.locate person_gid
#   # => #<Person:0x007fae94bf6298 @id="1">
#
# source://globalid//lib/global_id/identification.rb#28
module GlobalID::Identification
  # Returns the Global ID of the model.
  #
  #   model = Person.new id: 1
  #   global_id = model.to_global_id
  #   global_id.modal_class # => Person
  #   global_id.modal_id # => "1"
  #   global_id.to_param # => "Z2lkOi8vYm9yZGZvbGlvL1BlcnNvbi8x"
  #
  # source://globalid//lib/global_id/identification.rb#37
  def to_gid(options = T.unsafe(nil)); end

  # Returns the Global ID parameter of the model.
  #
  #   model = Person.new id: 1
  #   model.to_gid_param # => ""Z2lkOi8vYm9yZGZvbGlvL1BlcnNvbi8x"
  #
  # source://globalid//lib/global_id/identification.rb#46
  def to_gid_param(options = T.unsafe(nil)); end

  # Returns the Global ID of the model.
  #
  #   model = Person.new id: 1
  #   global_id = model.to_global_id
  #   global_id.modal_class # => Person
  #   global_id.modal_id # => "1"
  #   global_id.to_param # => "Z2lkOi8vYm9yZGZvbGlvL1BlcnNvbi8x"
  #
  # source://globalid//lib/global_id/identification.rb#37
  def to_global_id(options = T.unsafe(nil)); end

  # Returns the Signed Global ID of the model.
  # Signed Global IDs ensure that the data hasn't been tampered with.
  #
  #   model = Person.new id: 1
  #   signed_global_id = model.to_signed_global_id
  #   signed_global_id.modal_class # => Person
  #   signed_global_id.modal_id # => "1"
  #   signed_global_id.to_param # => "BAh7CEkiCGdpZAY6BkVUSSIiZ2..."
  #
  # ==== Expiration
  #
  # Signed Global IDs can expire some time in the future. This is useful if
  # there's a resource people shouldn't have indefinite access to, like a
  # share link.
  #
  #   expiring_sgid = Document.find(5).to_sgid(expires_in: 2.hours, for: 'sharing')
  #   # => #<SignedGlobalID:0x008fde45df8937 ...>
  #   # Within 2 hours...
  #   GlobalID::Locator.locate_signed(expiring_sgid.to_s, for: 'sharing')
  #   # => #<Document:0x007fae94bf6298 @id="5">
  #   # More than 2 hours later...
  #   GlobalID::Locator.locate_signed(expiring_sgid.to_s, for: 'sharing')
  #   # => nil
  #
  # In Rails, an auto-expiry of 1 month is set by default.
  #
  # You need to explicitly pass `expires_in: nil` to generate a permanent
  # SGID that will not expire,
  #
  #   never_expiring_sgid = Document.find(5).to_sgid(expires_in: nil)
  #   # => #<SignedGlobalID:0x008fde45df8937 ...>
  #
  #   # Any time later...
  #   GlobalID::Locator.locate_signed never_expiring_sgid
  #   # => #<Document:0x007fae94bf6298 @id="5">
  #
  # It's also possible to pass a specific expiry time
  #
  #   explicit_expiring_sgid = SecretAgentMessage.find(5).to_sgid(expires_at: Time.now.advance(hours: 1))
  #   # => #<SignedGlobalID:0x008fde45df8937 ...>
  #
  #   # 1 hour later...
  #   GlobalID::Locator.locate_signed explicit_expiring_sgid.to_s
  #   # => nil
  #
  # Note that an explicit `:expires_at` takes precedence over a relative `:expires_in`.
  #
  # ==== Purpose
  #
  # You can even bump the security up some more by explaining what purpose a
  # Signed Global ID is for. In this way evildoers can't reuse a sign-up
  # form's SGID on the login page. For example.
  #
  #   signup_person_sgid = Person.find(1).to_sgid(for: 'signup_form')
  #   # => #<SignedGlobalID:0x007fea1984b520
  #   GlobalID::Locator.locate_signed(signup_person_sgid.to_s, for: 'signup_form')
  #   => #<Person:0x007fae94bf6298 @id="1">
  #
  # source://globalid//lib/global_id/identification.rb#107
  def to_sgid(options = T.unsafe(nil)); end

  # Returns the Signed Global ID parameter.
  #
  #   model = Person.new id: 1
  #   model.to_sgid_param # => "BAh7CEkiCGdpZAY6BkVUSSIiZ2..."
  #
  # source://globalid//lib/global_id/identification.rb#116
  def to_sgid_param(options = T.unsafe(nil)); end

  # Returns the Signed Global ID of the model.
  # Signed Global IDs ensure that the data hasn't been tampered with.
  #
  #   model = Person.new id: 1
  #   signed_global_id = model.to_signed_global_id
  #   signed_global_id.modal_class # => Person
  #   signed_global_id.modal_id # => "1"
  #   signed_global_id.to_param # => "BAh7CEkiCGdpZAY6BkVUSSIiZ2..."
  #
  # ==== Expiration
  #
  # Signed Global IDs can expire some time in the future. This is useful if
  # there's a resource people shouldn't have indefinite access to, like a
  # share link.
  #
  #   expiring_sgid = Document.find(5).to_sgid(expires_in: 2.hours, for: 'sharing')
  #   # => #<SignedGlobalID:0x008fde45df8937 ...>
  #   # Within 2 hours...
  #   GlobalID::Locator.locate_signed(expiring_sgid.to_s, for: 'sharing')
  #   # => #<Document:0x007fae94bf6298 @id="5">
  #   # More than 2 hours later...
  #   GlobalID::Locator.locate_signed(expiring_sgid.to_s, for: 'sharing')
  #   # => nil
  #
  # In Rails, an auto-expiry of 1 month is set by default.
  #
  # You need to explicitly pass `expires_in: nil` to generate a permanent
  # SGID that will not expire,
  #
  #   never_expiring_sgid = Document.find(5).to_sgid(expires_in: nil)
  #   # => #<SignedGlobalID:0x008fde45df8937 ...>
  #
  #   # Any time later...
  #   GlobalID::Locator.locate_signed never_expiring_sgid
  #   # => #<Document:0x007fae94bf6298 @id="5">
  #
  # It's also possible to pass a specific expiry time
  #
  #   explicit_expiring_sgid = SecretAgentMessage.find(5).to_sgid(expires_at: Time.now.advance(hours: 1))
  #   # => #<SignedGlobalID:0x008fde45df8937 ...>
  #
  #   # 1 hour later...
  #   GlobalID::Locator.locate_signed explicit_expiring_sgid.to_s
  #   # => nil
  #
  # Note that an explicit `:expires_at` takes precedence over a relative `:expires_in`.
  #
  # ==== Purpose
  #
  # You can even bump the security up some more by explaining what purpose a
  # Signed Global ID is for. In this way evildoers can't reuse a sign-up
  # form's SGID on the login page. For example.
  #
  #   signup_person_sgid = Person.find(1).to_sgid(for: 'signup_form')
  #   # => #<SignedGlobalID:0x007fea1984b520
  #   GlobalID::Locator.locate_signed(signup_person_sgid.to_s, for: 'signup_form')
  #   => #<Person:0x007fae94bf6298 @id="1">
  #
  # source://globalid//lib/global_id/identification.rb#107
  def to_signed_global_id(options = T.unsafe(nil)); end
end

# source://globalid//lib/global_id/locator.rb#4
module GlobalID::Locator
  class << self
    # Takes either a GlobalID or a string that can be turned into a GlobalID
    #
    # Options:
    # * <tt>:includes</tt> - A Symbol, Array, Hash or combination of them.
    #   The same structure you would pass into a +includes+ method of Active Record.
    #   If present, locate will load all the relationships specified here.
    #   See https://guides.rubyonrails.org/active_record_querying.html#eager-loading-associations.
    # * <tt>:only</tt> - A class, module or Array of classes and/or modules that are
    #   allowed to be located.  Passing one or more classes limits instances of returned
    #   classes to those classes or their subclasses.  Passing one or more modules in limits
    #   instances of returned classes to those including that module.  If no classes or
    #   modules match, +nil+ is returned.
    #
    # source://globalid//lib/global_id/locator.rb#20
    def locate(gid, options = T.unsafe(nil)); end

    # Takes an array of GlobalIDs or strings that can be turned into a GlobalIDs.
    # All GlobalIDs must belong to the same app, as they will be located using
    # the same locator using its locate_many method.
    #
    # By default the GlobalIDs will be located using Model.find(array_of_ids), so the
    # models must respond to that finder signature.
    #
    # This approach will efficiently call only one #find (or #where(id: id), when using ignore_missing)
    # per model class, but still interpolate the results to match the order in which the gids were passed.
    #
    # Options:
    # * <tt>:includes</tt> - A Symbol, Array, Hash or combination of them
    #   The same structure you would pass into a includes method of Active Record.
    #   @see https://guides.rubyonrails.org/active_record_querying.html#eager-loading-associations
    #   If present, locate_many will load all the relationships specified here.
    #   Note: It only works if all the gids models have that relationships.
    # * <tt>:only</tt> - A class, module or Array of classes and/or modules that are
    #   allowed to be located.  Passing one or more classes limits instances of returned
    #   classes to those classes or their subclasses.  Passing one or more modules in limits
    #   instances of returned classes to those including that module.  If no classes or
    #   modules match, +nil+ is returned.
    # * <tt>:ignore_missing</tt> - By default, locate_many will call #find on the model to locate the
    #   ids extracted from the GIDs. In Active Record (and other data stores following the same pattern),
    #   #find will raise an exception if a named ID can't be found. When you set this option to true,
    #   we will use #where(id: ids) instead, which does not raise on missing records.
    #
    # source://globalid//lib/global_id/locator.rb#60
    def locate_many(gids, options = T.unsafe(nil)); end

    # Takes an array of SignedGlobalIDs or strings that can be turned into a SignedGlobalIDs.
    # The SignedGlobalIDs are located using Model.find(array_of_ids), so the models must respond to
    # that finder signature.
    #
    # This approach will efficiently call only one #find per model class, but still interpolate
    # the results to match the order in which the gids were passed.
    #
    # Options:
    # * <tt>:includes</tt> - A Symbol, Array, Hash or combination of them
    #   The same structure you would pass into a includes method of Active Record.
    #   @see https://guides.rubyonrails.org/active_record_querying.html#eager-loading-associations
    #   If present, locate_many_signed will load all the relationships specified here.
    #   Note: It only works if all the gids models have that relationships.
    # * <tt>:only</tt> - A class, module or Array of classes and/or modules that are
    #   allowed to be located.  Passing one or more classes limits instances of returned
    #   classes to those classes or their subclasses.  Passing one or more modules in limits
    #   instances of returned classes to those including that module.  If no classes or
    #   modules match, +nil+ is returned.
    #
    # source://globalid//lib/global_id/locator.rb#103
    def locate_many_signed(sgids, options = T.unsafe(nil)); end

    # Takes either a SignedGlobalID or a string that can be turned into a SignedGlobalID
    #
    # Options:
    # * <tt>:includes</tt> - A Symbol, Array, Hash or combination of them
    #   The same structure you would pass into a includes method of Active Record.
    #   @see https://guides.rubyonrails.org/active_record_querying.html#eager-loading-associations
    #   If present, locate_signed will load all the relationships specified here.
    # * <tt>:only</tt> - A class, module or Array of classes and/or modules that are
    #   allowed to be located.  Passing one or more classes limits instances of returned
    #   classes to those classes or their subclasses.  Passing one or more modules in limits
    #   instances of returned classes to those including that module.  If no classes or
    #   modules match, +nil+ is returned.
    #
    # source://globalid//lib/global_id/locator.rb#81
    def locate_signed(sgid, options = T.unsafe(nil)); end

    # Tie a locator to an app.
    # Useful when different apps collaborate and reference each others' Global IDs.
    #
    # The locator can be either a block or a class.
    #
    # Using a block:
    #
    #   GlobalID::Locator.use :foo do |gid, options|
    #     FooRemote.const_get(gid.model_name).find(gid.model_id)
    #   end
    #
    # Using a class:
    #
    #   GlobalID::Locator.use :bar, BarLocator.new
    #
    #   class BarLocator
    #     def locate(gid, options = {})
    #       @search_client.search name: gid.model_name, id: gid.model_id
    #     end
    #   end
    #
    # @raise [ArgumentError]
    #
    # source://globalid//lib/global_id/locator.rb#127
    def use(app, locator = T.unsafe(nil), &locator_block); end

    private

    # @return [Boolean]
    #
    # source://globalid//lib/global_id/locator.rb#140
    def find_allowed?(model_class, only = T.unsafe(nil)); end

    # source://globalid//lib/global_id/locator.rb#136
    def locator_for(gid); end

    # source://globalid//lib/global_id/locator.rb#148
    def normalize_app(app); end

    # source://globalid//lib/global_id/locator.rb#144
    def parse_allowed(gids, only = T.unsafe(nil)); end
  end
end

# source://globalid//lib/global_id/locator.rb#156
class GlobalID::Locator::BaseLocator
  # source://globalid//lib/global_id/locator.rb#157
  def locate(gid, options = T.unsafe(nil)); end

  # source://globalid//lib/global_id/locator.rb#165
  def locate_many(gids, options = T.unsafe(nil)); end

  private

  # source://globalid//lib/global_id/locator.rb#189
  def find_records(model_class, ids, options); end

  # @return [Boolean]
  #
  # source://globalid//lib/global_id/locator.rb#199
  def model_id_is_valid?(gid); end

  # source://globalid//lib/global_id/locator.rb#203
  def primary_key(model_class); end
end

# source://globalid//lib/global_id/locator.rb#228
class GlobalID::Locator::BlockLocator
  # @return [BlockLocator] a new instance of BlockLocator
  #
  # source://globalid//lib/global_id/locator.rb#229
  def initialize(block); end

  # source://globalid//lib/global_id/locator.rb#233
  def locate(gid, options = T.unsafe(nil)); end

  # source://globalid//lib/global_id/locator.rb#237
  def locate_many(gids, options = T.unsafe(nil)); end
end

# source://globalid//lib/global_id/locator.rb#226
GlobalID::Locator::DEFAULT_LOCATOR = T.let(T.unsafe(nil), GlobalID::Locator::UnscopedLocator)

# source://globalid//lib/global_id/locator.rb#5
class GlobalID::Locator::InvalidModelIdError < ::StandardError; end

# source://globalid//lib/global_id/locator.rb#208
class GlobalID::Locator::UnscopedLocator < ::GlobalID::Locator::BaseLocator
  # source://globalid//lib/global_id/locator.rb#209
  def locate(gid, options = T.unsafe(nil)); end

  private

  # source://globalid//lib/global_id/locator.rb#214
  def find_records(model_class, ids, options); end

  # source://globalid//lib/global_id/locator.rb#218
  def unscoped(model_class); end
end

# source://globalid//lib/global_id/railtie.rb#12
class GlobalID::Railtie < ::Rails::Railtie; end

# source://globalid//lib/global_id/verifier.rb#4
class GlobalID::Verifier < ::ActiveSupport::MessageVerifier
  private

  # source://globalid//lib/global_id/verifier.rb#10
  def decode(data, **_arg1); end

  # source://globalid//lib/global_id/verifier.rb#6
  def encode(data, **_arg1); end
end

# source://globalid//lib/global_id/signed_global_id.rb#4
class SignedGlobalID < ::GlobalID
  # @return [SignedGlobalID] a new instance of SignedGlobalID
  #
  # source://globalid//lib/global_id/signed_global_id.rb#59
  def initialize(gid, options = T.unsafe(nil)); end

  # source://globalid//lib/global_id/signed_global_id.rb#71
  def ==(other); end

  # Returns the value of attribute expires_at.
  #
  # source://globalid//lib/global_id/signed_global_id.rb#57
  def expires_at; end

  # source://globalid//lib/global_id/signed_global_id.rb#75
  def inspect; end

  # Returns the value of attribute purpose.
  #
  # source://globalid//lib/global_id/signed_global_id.rb#57
  def purpose; end

  # source://globalid//lib/global_id/signed_global_id.rb#66
  def to_param; end

  # source://globalid//lib/global_id/signed_global_id.rb#66
  def to_s; end

  # Returns the value of attribute verifier.
  #
  # source://globalid//lib/global_id/signed_global_id.rb#57
  def verifier; end

  private

  # source://globalid//lib/global_id/signed_global_id.rb#80
  def pick_expiration(options); end

  class << self
    # Returns the value of attribute expires_in.
    #
    # source://globalid//lib/global_id/signed_global_id.rb#8
    def expires_in; end

    # Sets the attribute expires_in
    #
    # @param value the value to set the attribute expires_in to.
    #
    # source://globalid//lib/global_id/signed_global_id.rb#8
    def expires_in=(_arg0); end

    # source://globalid//lib/global_id/signed_global_id.rb#10
    def parse(sgid, options = T.unsafe(nil)); end

    # source://globalid//lib/global_id/signed_global_id.rb#24
    def pick_purpose(options); end

    # Grab the verifier from options and fall back to SignedGlobalID.verifier.
    # Raise ArgumentError if neither is available.
    #
    # source://globalid//lib/global_id/signed_global_id.rb#16
    def pick_verifier(options); end

    # Returns the value of attribute verifier.
    #
    # source://globalid//lib/global_id/signed_global_id.rb#8
    def verifier; end

    # Sets the attribute verifier
    #
    # @param value the value to set the attribute verifier to.
    #
    # source://globalid//lib/global_id/signed_global_id.rb#8
    def verifier=(_arg0); end

    private

    # source://globalid//lib/global_id/signed_global_id.rb#50
    def raise_if_expired(expires_at); end

    # source://globalid//lib/global_id/signed_global_id.rb#29
    def verify(sgid, options); end

    # source://globalid//lib/global_id/signed_global_id.rb#40
    def verify_with_legacy_self_validated_metadata(sgid, options); end

    # source://globalid//lib/global_id/signed_global_id.rb#34
    def verify_with_verifier_validated_metadata(sgid, options); end
  end
end

# source://globalid//lib/global_id/signed_global_id.rb#5
class SignedGlobalID::ExpiredMessage < ::StandardError; end

# source://globalid//lib/global_id/uri/gid.rb#27
class URI::GID < ::URI::Generic
  # URI::GID encodes an app unique reference to a specific model as an URI.
  # It has the components: app name, model class name, model id and params.
  # All components except params are required.
  #
  # The URI format looks like "gid://app/model_name/model_id".
  #
  # Simple metadata can be stored in params. Useful if your app has multiple databases,
  # for instance, and you need to find out which one to look up the model in.
  #
  # Params will be encoded as query parameters like so
  # "gid://app/model_name/model_id?key=value&another_key=another_value".
  #
  # Params won't be typecast, they're always strings.
  # For convenience params can be accessed using both strings and symbol keys.
  #
  # Multi value params aren't supported. Any params encoding multiple values under
  # the same key will return only the last value. For example, when decoding
  # params like "key=first_value&key=last_value" key will only be last_value.
  #
  # Read the documentation for +parse+, +create+ and +build+ for more.
  #
  # source://uri/0.12.1/uri/generic.rb#243
  def app; end

  # source://globalid//lib/global_id/uri/gid.rb#107
  def deconstruct_keys(_keys); end

  # Returns the value of attribute model_id.
  #
  # source://globalid//lib/global_id/uri/gid.rb#29
  def model_id; end

  # Returns the value of attribute model_name.
  #
  # source://globalid//lib/global_id/uri/gid.rb#29
  def model_name; end

  # Returns the value of attribute params.
  #
  # source://globalid//lib/global_id/uri/gid.rb#29
  def params; end

  # source://globalid//lib/global_id/uri/gid.rb#102
  def to_s; end

  protected

  # Ruby 2.2 uses #query= instead of #set_query
  #
  # source://globalid//lib/global_id/uri/gid.rb#118
  def query=(query); end

  # source://globalid//lib/global_id/uri/gid.rb#129
  def set_params(params); end

  # source://globalid//lib/global_id/uri/gid.rb#112
  def set_path(path); end

  # Ruby 2.1 or less uses #set_query to assign the query
  #
  # source://globalid//lib/global_id/uri/gid.rb#124
  def set_query(query); end

  private

  # source://globalid//lib/global_id/uri/gid.rb#136
  def check_host(host); end

  # source://globalid//lib/global_id/uri/gid.rb#141
  def check_path(path); end

  # source://globalid//lib/global_id/uri/gid.rb#146
  def check_scheme(scheme); end

  # source://globalid//lib/global_id/uri/gid.rb#195
  def parse_query_params(query); end

  # source://globalid//lib/global_id/uri/gid.rb#154
  def set_model_components(path, validate = T.unsafe(nil)); end

  # @raise [URI::InvalidComponentError]
  #
  # source://globalid//lib/global_id/uri/gid.rb#174
  def validate_component(component); end

  # @raise [InvalidModelIdError]
  #
  # source://globalid//lib/global_id/uri/gid.rb#188
  def validate_model_id(model_id_part); end

  # @raise [MissingModelIdError]
  #
  # source://globalid//lib/global_id/uri/gid.rb#181
  def validate_model_id_section(model_id, model_name); end

  class << self
    # Create a new URI::GID from components with argument check.
    #
    # The allowed components are app, model_name, model_id and params, which can be
    # either a hash or an array.
    #
    # Using a hash:
    #
    #   URI::GID.build(app: 'bcx', model_name: 'Person', model_id: '1', params: { key: 'value' })
    #
    # Using an array, the arguments must be in order [app, model_name, model_id, params]:
    #
    #   URI::GID.build(['bcx', 'Person', '1', key: 'value'])
    #
    # source://globalid//lib/global_id/uri/gid.rb#88
    def build(args); end

    # Shorthand to build a URI::GID from an app, a model and optional params.
    #
    #   URI::GID.create('bcx', Person.find(5), database: 'superhumans')
    #
    # source://globalid//lib/global_id/uri/gid.rb#72
    def create(app, model, params = T.unsafe(nil)); end

    # Create a new URI::GID by parsing a gid string with argument check.
    #
    #   URI::GID.parse 'gid://bcx/Person/1?key=value'
    #
    # This differs from URI() and URI.parse which do not check arguments.
    #
    #   URI('gid://bcx')             # => URI::GID instance
    #   URI.parse('gid://bcx')       # => URI::GID instance
    #   URI::GID.parse('gid://bcx/') # => raises URI::InvalidComponentError
    #
    # source://globalid//lib/global_id/uri/gid.rb#64
    def parse(uri); end

    # Validates +app+'s as URI hostnames containing only alphanumeric characters
    # and hyphens. An ArgumentError is raised if +app+ is invalid.
    #
    #   URI::GID.validate_app('bcx')     # => 'bcx'
    #   URI::GID.validate_app('foo-bar') # => 'foo-bar'
    #
    #   URI::GID.validate_app(nil)       # => ArgumentError
    #   URI::GID.validate_app('foo/bar') # => ArgumentError
    #
    # source://globalid//lib/global_id/uri/gid.rb#48
    def validate_app(app); end
  end
end

# source://globalid//lib/global_id/uri/gid.rb#134
URI::GID::COMPONENT = T.let(T.unsafe(nil), Array)

# source://globalid//lib/global_id/uri/gid.rb#37
URI::GID::COMPOSITE_MODEL_ID_DELIMITER = T.let(T.unsafe(nil), String)

# Maximum size of a model id segment
#
# source://globalid//lib/global_id/uri/gid.rb#36
URI::GID::COMPOSITE_MODEL_ID_MAX_SIZE = T.let(T.unsafe(nil), Integer)

# source://globalid//lib/global_id/uri/gid.rb#33
class URI::GID::InvalidModelIdError < ::URI::InvalidComponentError; end

# Raised when creating a Global ID for a model without an id
#
# source://globalid//lib/global_id/uri/gid.rb#32
class URI::GID::MissingModelIdError < ::URI::InvalidComponentError; end