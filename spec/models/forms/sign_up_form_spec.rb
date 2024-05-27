# typed: false
# frozen_string_literal: true

describe Forms::SignUpForm do
  let(:email) { "example@example.com" }
  let(:back) { "/foo/bar" }
  let(:form) do
    Forms::SignUpForm.new(
      email: email,
      back: back
    )
  end

  describe "#valid?" do
    context "when attributes are valid" do
      it do
        expect(form.valid?).to be true
        expect(form.errors.full_messages).to eq []
      end
    end

    context "when attributes are invalid" do
      context "email" do
        context "is invalid format" do
          let(:email) { "not-email" }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq ["メールアドレスは不正な値です"]
          end
        end
      end

      context "back" do
        context "is invalid format" do
          let(:back) { "https://example.com/foo/bar" }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq ["戻り先は不正な値です"]
          end
        end
      end
    end
  end
end
