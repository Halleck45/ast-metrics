package namer

import "testing"

func TestFallbackFromNamespaces(t *testing.T) {
	tests := []struct {
		name       string
		classNames []string
		want       string
	}{
		{
			name:       "generic penultimate segment shared by all classes",
			classNames: []string{`App\Http\UserController`, `App\Http\OrderController`, `App\Http\ProductController`},
			want:       "",
		},
		{
			name:       "non-generic penultimate segment shared by all classes",
			classNames: []string{`App\Payment\Stripe`, `App\Payment\Braintree`, `App\Payment\Paypal`},
			want:       "Payment",
		},
		{
			name:       "mixed penultimate segments returns most common",
			classNames: []string{`Guzzle\Http\Client`, `Guzzle\Http\Message`, `Guzzle\Cookie\Jar`},
			want:       "Http",
		},
		{
			name:       "fewer than 3 segments returns empty",
			classNames: []string{`Foo\Bar`, `Baz\Qux`},
			want:       "",
		},
		{
			name:       "single class returns empty",
			classNames: []string{`App\Http\Client`},
			want:       "",
		},
		{
			name:       "generic segment but not all classes share it",
			classNames: []string{`App\Http\Client`, `App\Http\Message`, `Other\Payment\Stripe`},
			want:       "Http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fallbackFromNamespaces(tt.classNames)
			if got != tt.want {
				t.Errorf("fallbackFromNamespaces() = %q, want %q", got, tt.want)
			}
		})
	}
}
