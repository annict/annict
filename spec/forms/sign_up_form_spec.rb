# frozen_string_literal: true

describe SignUpForm do
  let(:email) { "example@example.com" }
  let(:back) { "/foo/bar" }
  let(:form) do
    SignUpForm.new(
      email: email,
      back: back
    )
  end

  describe "#valid?" do
    context "when attributes are valid" do
      it do
        expect(form.valid?).to be true
        expect(form.error_messages).to eq []
      end
    end

    context "when attributes are invalid" do
      context "email" do
        context "is invalid format" do
          let(:email) { "not-email" }

          it do
            expect(form.valid?).to be false
            expect(form.error_messages).to eq ["メールアドレスが不正です"]
          end
        end
      end

      context "back" do
        context "is invalid format" do
          let(:back) { "https://example.com/foo/bar" }

          it do
            expect(form.valid?).to be false
            expect(form.error_messages).to eq ["戻り先が不正です"]
          end
        end
      end
    end
  end
end
