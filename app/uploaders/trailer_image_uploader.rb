# typed: false
# frozen_string_literal: true

class TrailerImageUploader < ImageUploaderBase
  Attacher.validate do
    # Empty braces are required
    # https://github.com/shrinerb/shrine/blob/8aa88ea1a64ea1a694e7a9094f03ce6a38119378/doc/validation.md#inheritance
    super()
  end
end
