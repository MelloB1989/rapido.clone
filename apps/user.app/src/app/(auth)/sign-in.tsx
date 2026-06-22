import { useRouter } from 'expo-router';
import { useState } from 'react';
import { Alert, StyleSheet } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { AuthTextField } from '@/components/auth/auth-text-field';
import { PrimaryButton } from '@/components/auth/primary-button';
import { ThemedText } from '@/components/themed-text';
import { ThemedView } from '@/components/themed-view';
import { Spacing } from '@/constants/theme';
import { UnconfirmedUserError, useAuth } from '@/hooks/use-auth';

export default function SignInScreen() {
  const router = useRouter();
  const { signIn } = useAuth();
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);

  const onSubmit = async () => {
    if (!phone || !password) {
      Alert.alert('Missing details', 'Please enter your phone number and password.');
      return;
    }
    setLoading(true);
    try {
      await signIn(phone, password);
    } catch (err) {
      if (err instanceof UnconfirmedUserError) {
        router.push({ pathname: '/confirm', params: { phone: err.phone } });
        return;
      }
      Alert.alert('Sign in failed', err instanceof Error ? err.message : 'Something went wrong.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <ThemedView style={styles.container}>
      <SafeAreaView style={styles.safeArea}>
        <ThemedText type="title" style={styles.title}>
          Welcome back
        </ThemedText>
        <ThemedView type="backgroundElement" style={styles.form}>
          <AuthTextField
            label="Phone number"
            value={phone}
            onChangeText={setPhone}
            keyboardType="phone-pad"
            placeholder="98765 43210"
          />
          <AuthTextField label="Password" value={password} onChangeText={setPassword} secureTextEntry />
          <PrimaryButton title="Sign in" onPress={onSubmit} loading={loading} />
        </ThemedView>
        <ThemedText type="link" style={styles.footer} onPress={() => router.replace('/sign-up')}>
          New here? <ThemedText type="linkPrimary">Create an account</ThemedText>
        </ThemedText>
      </SafeAreaView>
    </ThemedView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
  safeArea: { flex: 1, paddingHorizontal: Spacing.four, gap: Spacing.four, justifyContent: 'center' },
  title: { textAlign: 'center' },
  form: { gap: Spacing.three, padding: Spacing.four, borderRadius: Spacing.four },
  footer: { textAlign: 'center' },
});
